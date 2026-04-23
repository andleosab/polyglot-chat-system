import { betterAuth } from "better-auth";
import { Pool } from "pg";
import { env } from "$env/dynamic/private";
import { getRequestEvent } from "$app/server";
import { sveltekitCookies } from "better-auth/svelte-kit";
import { issueServiceToken } from "./server/jwt";

import {
  GOOGLE_CLIENT_ID,
  GOOGLE_CLIENT_SECRET
} from "$env/static/private";

export const auth = betterAuth({
	baseURL: env.BETTER_AUTH_URL!,
	secret: env.BETTER_AUTH_SECRET!,
  database: new Pool({
    connectionString: env.DATABASE_URL,
  }),
  user: {
    additionalFields: {
      username: {
        type: "string",
        required: false,      // required at signup
        // defaultValue: "",     // default if not provided
      },
	  useruuid: {
		type: "string",
		required: false,      // required at signup
		// defaultValue: crypto.randomUUID().toString(),     // default if not provided
	  },
    },
  },  
  databaseHooks: {
	user: {
		create: {
            before: async (user, ctx) => {
				console.log("Creating user:", user, " with ctx: ", ctx?.body);
                return {
                    data: {
                        ...user,
                        username: user.username ?? user.email?.split("@")[0] ?? user.email,
                        useruuid: crypto.randomUUID().toString() 
                    }
                };
            },			
			after: async (user, ctx ) => {

				// const username =
				// typeof user.username === "string"
				// 	? user.username
				// 	: user.email;				

				const token = issueServiceToken('chat-user-service', {
					userId: user.useruuid as string,
					email: user.email,
					username: user.username as string,
				});
				console.log("User created:", user);
				try {
					const response = await fetch(`${env.USER_API_BASE}/api/users`, {
						method: "POST",
						headers: {
							"Content-Type": "application/json",
							"Authorization": `Bearer ${token}`
						},
						body: JSON.stringify({
							username: user.username,
							email: user.email,
							userid: user.useruuid,  
						}),
					});
					if (!response.ok) {
						const errBody = await response.json().catch(() => null);
						throw {
							status: response.status,
							body: errBody
						};    
					}
					console.log("User synchronized:", user.username, user.useruuid);
					
				} catch (error) {
					console.error("Error creating user:", error);
					throw error ;
				}
			}
	  	}
	}
  },

  plugins: [sveltekitCookies(getRequestEvent)],	

	socialProviders: {
		google: {
			clientId: env.GOOGLE_CLIENT_ID || GOOGLE_CLIENT_ID ,
			clientSecret: env.GOOGLE_CLIENT_SECRET || GOOGLE_CLIENT_SECRET,
			accessType: "offline", 
        	prompt: "select_account consent", 
			scope: ["openid email profile"],
			// default, can be ommitted
			// redirectURI: `${env.BETTER_AUTH_URL}/auth/google/callback`,  
			authorizationParams: {
				login_hint: "user@example.com", // optional, pre-fill the email field on the Google sign-in page
				// scope: "openid email profile",
				// prompt: "consent",   // forces consent screen every time
			},			
		},
	},
	emailAndPassword: {
		enabled: true,
		async sendResetPassword(url, user) {
			console.log("Reset password url:", url);
		},
	},
	emailVerification: {
		sendOnSignUp: false, // TODO enable this option to send email to the user on sign up
		// sendVerificationEmail: async ({ user, url, token }, request) => {
		// 	// TODO add function(s) to send verification email.
		// },
	},
});