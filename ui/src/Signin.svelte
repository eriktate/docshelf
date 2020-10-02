<script lang="typescript">
	import { onMount } from "svelte";
	import { navigate } from "svelte-routing";
	import { login } from "./api";

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

	onMount(async () => {
		window.gapi.signin2.render("g-signin2");
	});
</script>


<div id="g-signin2" class="g-signin2" data-onsuccess="onSignIn"></div>
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
