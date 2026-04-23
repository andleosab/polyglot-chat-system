import { json } from "@sveltejs/kit";
import { issueWsToken } from "$lib/server/jwt";

// provides a WebSocket token for the authenticated user to connect to the chat delivery service
export const POST = async ({ locals }) => {
  if (!locals.currentUser) return new Response("unauthorized", { status: 401 });

  const token = issueWsToken("chat-delivery-service", locals.currentUser);
  return json({ token });
};