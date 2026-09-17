export interface LoginResponse {
  status: string;
  error?: string;
}

export interface MeResponse {
  username?: string;
  error?: string;
}

export async function login(username: string, password: string): Promise<LoginResponse> {
  const res = await fetch('/api/login', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ username, password }),
  });
  return res.json();
}

export async function logout(): Promise<void> {
  await fetch('/api/logout', { method: 'POST' });
}

export async function me(): Promise<MeResponse> {
  const res = await fetch('/api/me');
  return res.json();
}
