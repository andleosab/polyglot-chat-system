<script lang="ts">
import { signIn } from "$lib/auth-client";
import { writable } from "svelte/store";
import { Button } from "$lib/components/ui/button";
import * as Card from "$lib/components/ui/card";
import { Input } from "$lib/components/ui/input";
import { Label } from "$lib/components/ui/label";

const email = writable("");
const password = writable("");
const errorMessage = writable("");

const handleSignIn = async () => {
  await signIn.email(
    {
      email: $email,
      password: $password,
      callbackURL: "/",
    },
    {
      onError(context) {
        if (context.error.status >= 500) {
          errorMessage.set("Google sign-in is likely misconfigured. Check GOOGLE_CLIENT_ID and GOOGLE_CLIENT_SECRET.");
        } else {
          errorMessage.set(context.error.message);
        }
      },
    },
  );
};
</script>

<div class="flex min-h-screen items-center justify-center">
<Card.Root class="mx-auto w-full max-w-sm shadow-lg">
  <Card.Header>
    <Card.Title class="text-2xl">Login</Card.Title>
    <Card.Description>
      Enter your email below to login to your account
    </Card.Description>
  </Card.Header>
  <Card.Content>
    <div class="grid gap-4">
      <!-- Error message above email input -->
      {#if $errorMessage}
        <p class="text-sm text-red-600">{$errorMessage}</p>
      {/if}    
      <div class="grid gap-2">
        <Label for="email">Email</Label>
        <Input
          id="email"
          type="email"
          placeholder="m@example.com"
          required
          bind:value={$email}
        />
      </div>
      <div class="grid gap-2">
        <div class="flex items-center">
          <Label for="password">Password</Label>
          <a
            href="/forget-password"
            class="ml-auto inline-block text-sm underline"
          >
            Forgot your password?
          </a>
        </div>
        <Input id="password" type="password" required bind:value={$password} />
      </div>
      <Button type="button" class="w-full" onclick={handleSignIn}>Login</Button
      >
      <Button
        variant="outline"
        class="w-full"
        onclick={async () => {
          await signIn.social({
            provider: "google",
            callbackURL: "/chats",
          },
          {
            onError(context) {
              if (context.error.status >= 500) {
                errorMessage.set("Something went wrong. Please try again.");
              } else {
                errorMessage.set(context.error.message);
              }
            },
          }
        );
        }}>Login with Google</Button
      >
    </div>
    <div class="mt-4 text-center text-sm">
      Don&apos;t have an account?
      <a href="/sign-up" class="underline">Sign up</a>
    </div>
  </Card.Content>
</Card.Root>
</div>


   
  