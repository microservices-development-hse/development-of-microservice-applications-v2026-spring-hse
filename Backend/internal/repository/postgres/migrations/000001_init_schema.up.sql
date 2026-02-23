CREATE TABLE "Project" (
    "id" SERIAL PRIMARY KEY,
    "key" VARCHAR(10) UNIQUE NOT NULL,
    "title" TEXT NOT NULL
);

CREATE TABLE "Author" (
    "id" SERIAL PRIMARY KEY,
    "name" TEXT NOT NULL
);

CREATE TABLE "Issue" (
    "id" SERIAL PRIMARY KEY,
    "project_id" INTEGER NOT NULL REFERENCES "Project"("id") ON DELETE CASCADE,
    "author_id" INTEGER NOT NULL REFERENCES "Author"("id"),
    "assignee_id" INTEGER NOT NULL REFERENCES "Author"("id"),
    "key" TEXT NOT NULL UNIQUE,
    "summary" TEXT NOT NULL,
    "description" TEXT,
    "type" TEXT,
    "priority" TEXT,
    "status" TEXT,
    "created_time" TIMESTAMP WITH TIME ZONE,
    "closed_time" TIMESTAMP WITH TIME ZONE,
    "updated_time" TIMESTAMP WITH TIME ZONE,
    "time_spent" INTEGER DEFAULT 0
);

CREATE TABLE "StatusChanges" (
    "issue_id" INTEGER NOT NULL REFERENCES "Issue"("id") ON DELETE CASCADE,
    "author_id" INTEGER NOT NULL REFERENCES "Author"("id"),
    "change_time" TIMESTAMP WITH TIME ZONE NOT NULL,
    "from_status" TEXT,
    "to_status" TEXT
);

CREATE TABLE "OpenTaskTime" (
    "project_id" INTEGER NOT NULL REFERENCES "Project"("id") ON DELETE CASCADE,
    "creation_time" TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    "data" JSONB NOT NULL,
    PRIMARY KEY ("project_id")
);

CREATE TABLE "TaskStateTime" (
    "project_id" INTEGER NOT NULL REFERENCES "Project"("id") ON DELETE CASCADE,
    "creation_time" TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    "state" TEXT NOT NULL,
    "data" JSONB NOT NULL,
    PRIMARY KEY ("project_id", "state")
);

CREATE TABLE "ComplexityTaskTime" (
    "project_id" INTEGER NOT NULL REFERENCES "Project"("id") ON DELETE CASCADE,
    "creation_time" TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    "data" JSONB NOT NULL,
    PRIMARY KEY ("project_id")
);

CREATE TABLE "TaskPriorityCount" (
    "project_id" INTEGER NOT NULL REFERENCES "Project"("id") ON DELETE CASCADE,
    "creation_time" TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    "state" TEXT NOT NULL,
    "data" JSONB NOT NULL,
    PRIMARY KEY ("project_id", "state")
);

CREATE TABLE "ActivityByTask" (
    "project_id" INTEGER NOT NULL REFERENCES "Project"("id") ON DELETE CASCADE,
    "creation_time" TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    "state" TEXT NOT NULL,
    "data" JSONB NOT NULL,
    PRIMARY KEY ("project_id", "state")
);