import { Request as IttyRequest } from "itty-router";

export type IttyRequest = IttyRequest;

export type MethodType =
  | "GET"
  | "POST"
  | "PUT"
  | "DELETE"
  | "PATCH"
  | "HEAD"
  | "OPTIONS";

export interface IRequest extends IttyRequest {
  method: MethodType; // method is required to be on the interface
  url: string; // url is required to be on the interface
  optional?: string;
}

export interface IMethods extends IHTTPMethods {
  get: Route;
  post: Route;
  put: Route;
  delete: Route;
  patch: Route;
  head: Route;
  options: Route;
}

export type Handler = (
  req: IttyRequest,
  env: Env,
  ctx: ExecutionContext
) => Promise<Response>;

export interface Env {
	BUCKET: R2Bucket;
}
