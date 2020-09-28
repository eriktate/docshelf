<script lang="typescript">
	import { navigate } from "svelte-routing";
	import { login } from "./api.ts";

	export let success: () => void;
	let username: string = "";
	let password: string = "";

	async function handleSubmit() {
		try {
			await login(username, password);
			navigate("/", { replace: false });
			await success();
		} catch (err) {
			console.log("failed to login");
		}
	}
</script>


<div class="g-signin2" data-onsuccess="onSignIn"></div>
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
