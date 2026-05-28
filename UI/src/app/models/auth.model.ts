export interface ILoginRequest {
  email: string;
  password: string;
}

export interface ILoginResponse {
  token: string;
  email: string;
  expiresAt: number;
}

export interface IRegisterResponse {
  message: string;
}

export interface IAuthSession {
  token: string;
  email: string;
  expiresAt: number;
}

