import { createAuthClient } from "better-auth/svelte"; 
import { inferAdditionalFields } from "better-auth/client/plugins";
import type { auth } from "./auth";

export const client = createAuthClient({
  // you can pass client configuration here
//   baseURL: "http://localhost:5173",
  plugins: [inferAdditionalFields<typeof auth>()],
});
export const { signIn, signUp, useSession } = client;