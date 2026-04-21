package grpc

import (
	"context"
	"fmt"

	etlprocess "github.com/microservices-development-hse/connector/internal/etl-process"
	pb "github.com/microservices-development-hse/connector/internal/generated/connector"
	"github.com/microservices-development-hse/connector/internal/logger"
	dbmodels "github.com/microservices-development-hse/connector/internal/models/db"
)

type ConnectorGRPCServer struct {
	pb.UnimplementedConnectorServiceServer
	extractor *etlprocess.Extractor
	loader    *etlprocess.Loader
}

func NewConnectorGRPCServer(extractor *etlprocess.Extractor, loader *etlprocess.Loader) *ConnectorGRPCServer {
	return &ConnectorGRPCServer{
		extractor: extractor,
		loader:    loader,
	}
}

func (s *ConnectorGRPCServer) FetchRemoteProjects(ctx context.Context, _ *pb.Empty) (*pb.ProjectList, error) {
	projects, err := s.extractor.GetProjects()
	if err != nil {
		logger.Error("grpc: FetchRemoteProjects failed: %v", err)
		return nil, fmt.Errorf("failed to fetch projects: %w", err)
	}

	var pbProjects []*pb.Project
	for _, p := range projects {
		pbProjects = append(pbProjects, &pb.Project{
			Key:   p.Key,
			Title: p.Name,
			Url:   p.Self,
		})
	}

	return &pb.ProjectList{Projects: pbProjects}, nil
}

func (s *ConnectorGRPCServer) TriggerProjectImport(ctx context.Context, req *pb.ImportRequest) (*pb.ImportResponse, error) {
	projectKey := req.GetProjectKey()
	if projectKey == "" {
		return &pb.ImportResponse{
			Success: false,
			Message: "project_key is required",
		}, nil
	}

	logger.Info("grpc: TriggerProjectImport for project %q", projectKey)

	jiraIssues, err := s.extractor.GetAllIssues(ctx, projectKey)
	if err != nil {
		logger.Error("grpc: extract issues failed: %v", err)
		return &pb.ImportResponse{Success: false, Message: fmt.Sprintf("extract failed: %v", err)}, nil
	}

	jiraProjects, err := s.extractor.GetProjects()
	if err != nil {
		return &pb.ImportResponse{Success: false, Message: fmt.Sprintf("fetch projects failed: %v", err)}, nil
	}

	var projectID int

	for _, jp := range jiraProjects {
		if jp.Key != projectKey {
			continue
		}

		dbProject, err := etlprocess.TransformProject(jp)
		if err != nil {
			return &pb.ImportResponse{Success: false, Message: fmt.Sprintf("transform project failed: %v", err)}, nil
		}

		projectID, err = s.loader.LoadProject(ctx, dbProject)
		if err != nil {
			return &pb.ImportResponse{Success: false, Message: fmt.Sprintf("load project failed: %v", err)}, nil
		}

		break
	}

	if projectID == 0 {
		return &pb.ImportResponse{Success: false, Message: fmt.Sprintf("project %q not found in Jira", projectKey)}, nil
	}

	seenAuthors := make(map[string]dbmodels.Author)

	for _, ji := range jiraIssues {
		if ji.Fields.Creator.Name != "" {
			seenAuthors[ji.Fields.Creator.Name] = dbmodels.Author{
				ExternalID: ji.Fields.Creator.Name,
				Username:   ji.Fields.Creator.DisplayName,
			}
		}

		if ji.Fields.Assignee != nil && ji.Fields.Assignee.Name != "" {
			seenAuthors[ji.Fields.Assignee.Name] = dbmodels.Author{
				ExternalID: ji.Fields.Assignee.Name,
				Username:   ji.Fields.Assignee.DisplayName,
			}
		}

		if ji.Changelog != nil {
			for _, h := range ji.Changelog.Histories {
				if h.Author.Name != "" {
					seenAuthors[h.Author.Name] = dbmodels.Author{
						ExternalID: h.Author.Name,
						Username:   h.Author.DisplayName,
					}
				}
			}
		}
	}

	authorIDs, err := s.loader.UpsertAuthors(ctx, seenAuthors)
	if err != nil {
		return &pb.ImportResponse{Success: false, Message: fmt.Sprintf("upsert authors failed: %v", err)}, nil
	}

	var allIssues []dbmodels.Issue

	for _, ji := range jiraIssues {
		var authorID *int
		if id := authorIDs[ji.Fields.Creator.Name]; id != 0 {
			authorID = &id
		}

		var assigneeID *int

		if ji.Fields.Assignee != nil && ji.Fields.Assignee.Name != "" {
			if id := authorIDs[ji.Fields.Assignee.Name]; id != 0 {
				assigneeID = &id
			}
		}

		issue, err := etlprocess.TransformIssue(ji, projectID, authorID, assigneeID)
		if err != nil {
			return &pb.ImportResponse{Success: false, Message: fmt.Sprintf("transform issue failed: %v", err)}, nil
		}

		allIssues = append(allIssues, issue)
	}

	issueIDs, err := s.loader.LoadIssues(ctx, allIssues)
	if err != nil {
		return &pb.ImportResponse{Success: false, Message: fmt.Sprintf("load issues failed: %v", err)}, nil
	}

	var allStatusChanges []dbmodels.StatusChange

	for _, ji := range jiraIssues {
		if ji.Changelog == nil {
			continue
		}

		issueID, ok := issueIDs[ji.Key]
		if !ok {
			continue
		}

		changes := etlprocess.TransformStatusChanges(ji.Changelog, issueID, authorIDs)
		allStatusChanges = append(allStatusChanges, changes...)
	}

	if err := s.loader.LoadStatusChanges(ctx, allStatusChanges); err != nil {
		return &pb.ImportResponse{Success: false, Message: fmt.Sprintf("load status changes failed: %v", err)}, nil
	}

	logger.Info("grpc: project %q imported: %d issues, %d status changes", projectKey, len(allIssues), len(allStatusChanges))

	return &pb.ImportResponse{
		Success: true,
		Message: fmt.Sprintf("imported %d issues", len(allIssues)),
	}, nil
}
