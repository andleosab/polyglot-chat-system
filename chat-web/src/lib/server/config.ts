import { env } from '$env/dynamic/private';

function required(name: string): string {
  const value = env[name];
  if (!value) {
    throw new Error(`Missing required environment variable: ${name}`);
  }
  return value;
}

export const API = {
  users: required('USER_API_BASE'),
  messages: required('MESSAGE_API_BASE')
} as const;
