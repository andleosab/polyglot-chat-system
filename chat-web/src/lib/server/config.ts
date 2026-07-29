import { env } from '$env/dynamic/private';

function required(name: string): string {
  const value = env[name];
  if (!value) {
    throw new Error(`Missing required environment variable: ${name}`);
  }
  return value;
}

export const API = {
  get users() { return required('USER_API_BASE'); },
  get messages() { return required('MESSAGE_API_BASE'); },
  get presence() { return required('PRESENCE_API_BASE'); },
} as const;
