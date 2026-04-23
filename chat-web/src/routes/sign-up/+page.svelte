<script lang="ts">
import { goto } from "$app/navigation";
import { signUp } from "$lib/auth-client";
import { writable } from "svelte/store";
import { Button } from "$lib/components/ui/button";
import * as Card from "$lib/components/ui/card";
import { Input } from "$lib/components/ui/input";
import { Label } from "$lib/components/ui/label";
    import { error } from "@sveltejs/kit";

// Create writable stores for form fields
const firstName = writable("");
const lastName = writable("");
const email = writable("");
const password = writable("");
const username = writable(""); 
const errorMessage = writable("");

// Function to handle form submission
const handleSignUp = async () => {
	const user = {
		firstName: $firstName,
		lastName: $lastName,
		email: $email,
		password: $password,
    username: $username,
    uuid: crypto.randomUUID().toString(),
	};
	await signUp.email({
		email: user.email,
		password: user.password,
		name: `${user.firstName} ${user.lastName}`,
    username: user.username,
    useruuid: user.uuid.toString(),
		callbackURL: "/",
		fetchOptions: {
			onSuccess() {
				// alert("Your account has been created.");
				goto("/sign-in");
			},
			onError(context) {
          errorMessage.set(context.error.message);
			},
		},
	});
};
</script>

<div class="flex min-h-screen items-center justify-center">
<Card.Root class="mx-auto max-w-sm w-full shadow-sm">
  <Card.Header>
    <Card.Title class="text-xl">Sign Up</Card.Title>
    <Card.Description>
      Enter your information to create an account
    </Card.Description>
  </Card.Header>
  <Card.Content>
    <div class="grid gap-4">
          <!-- Error message above email input -->
      {#if $errorMessage}
        <p class="text-sm text-red-600">{$errorMessage}</p>
      {/if}    
      <div class="grid grid-cols-2 gap-4">
        <div class="grid gap-2">
          <Label for="first-name">First name</Label>
          <Input
            id="first-name"
            placeholder="Max"
            required
            bind:value={$firstName}
          />
        </div>
        <div class="grid gap-2">
          <Label for="last-name">Last name</Label>
          <Input
            id="last-name"
            placeholder="Robinson"
            required
            bind:value={$lastName}
          />
        </div>
      </div>
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
        <Label for="username">Username</Label>
        <Input
          id="username"
          type="text"
          name="username"
          placeholder="john_doe"
          required
          bind:value={$username}
        />
      </div>  


      <div class="grid gap-2">
        <Label for="password">Password</Label>
        <Input id="password" type="password" bind:value={$password} />
      </div>
      <Button type="button" class="w-full" onclick={handleSignUp}
        >Create an account</Button
      >
      <Button variant="outline" class="w-full">Sign up with GitHub</Button>
    </div>
    <div class="mt-4 text-center text-sm">
      Already have an account?
      <a href="/sign-in" class="underline"> Sign in </a>
    </div>
  </Card.Content>
</Card.Root>
</div>