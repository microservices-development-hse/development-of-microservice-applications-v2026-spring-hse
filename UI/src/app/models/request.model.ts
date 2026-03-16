import { Links } from "./links.model";
import { PageInfo } from "./pageInfo.model";
import { IProj } from "./proj.model";

export interface IRequest {
  _links?: Links;
  projects: IProj[];
  message?: string;
  name?: string;
  pageInfo: PageInfo;
  status?: boolean;
}
