<script lang="typescript">
	import { onMount } from "svelte";
	import { navigate } from "svelte-routing";
	import { login } from "./api";
	import retry from "./retry";

	export let success: () => void;
	let username: string = "";
	let password: string = "";

	async function handleSubmit() {
		try {
			await login(username, password);
			navigate("/", { replace: false });
			success();
		} catch (err) {
			console.log("failed to login");
		}
	}

	async function handleGoogleSignin(user: any): Promise<void> {
	  console.log("Signing in");
		const idToken = user.getAuthResponse().id_token;
		try {
			await login("", idToken); // login with no email assumes oauth
			user.disconnect();
			navigate("/", { replace: false });
		} catch (err) {
			console.log("failed to sign in with oauth");
		}
	}

	function handleGoogleFailure(err: any): void {
		console.log("Failed signin: ", err);
	}

	onMount(async () => {
		retry(async () => window.gapi.signin2.render("google-signin", {
			onsuccess: handleGoogleSignin,
			onfailure: handleGoogleFailure,
		}));
	});
</script>


<div id="google-signin"></div>
<form on:submit|preventDefault={handleSubmit}>
	<div>
		<input type="text" placeholder="Username" bind:value={username} />
	</div>
	<div>
		<input type="password" placeholder="Password" bind:value={password} />
	</div>
	<div>
		<button type="submit">Sign In</button>
	</div>
</form>
