<script lang="typescript">
	import { onMount } from "svelte";
	import { navigate } from "svelte-routing";
	import { login } from "./api";

	export let success: () => Promise<void>;
	let username: string = "";
	let password: string = "";

	const githubClientId: string = "64e0e1015fd8eaac389c";
	async function handleSubmit() {
		try {
			await login(username, password);
			navigate("/", { replace: false });
			success();
		} catch (err) {
			console.log("failed to login");
		}
	}

	async function handleGoogleSignin(): Promise<void> {
		const auth = window.gapi.auth2.getAuthInstance();
		const user = await auth.signIn();
		const idToken = user.getAuthResponse().id_token;
		try {
			await login("", idToken); // login with no email assumes oauth
			user.disconnect();
			success();
			navigate("/", { replace: false });
		} catch (err) {
			console.log("failed to sign in with oauth");
		}
	}

	async function handleGithubSignin(): Promise<void> {
		window.location.href = `https://github.com/login/oauth/authorize?client_id=${githubClientId}&scope=user:email&redirect_uri=http://localhost:9001/oauth/github`
	}

	onMount(async () => {
		window.gapi.load("auth2", function() {
			window.gapi.auth2.init();
		})
	});
</script>


<div class="pane">
	<h3>Sign In</h3>
	<form on:submit|preventDefault={handleSubmit}>
		<div class="oauth-buttons">
			<button type="button" id="google-signin" on:click={handleGoogleSignin}>
				<img class="icon" src="/assets/icons/google.png" alt="Google Sign In" />
				<span>Sign In With Google</span>
			</button>
			<button type="button" id="github-signin" on:click={handleGithubSignin}>
				<img class="icon" src="/assets/icons/github.png" alt="Github Sign In" />
				<span>Sign In With Github</span>
			</button>
		</div>
		<div>
			<input type="text" placeholder="Username" bind:value={username} />
		</div>
		<div>
			<input type="password" placeholder="Password" bind:value={password} />
		</div>
		<div>
			<button type="submit" class="submit">Sign In</button>
		</div>
	</form>
</div>

<style lang="scss">
	.pane {
		margin: auto;
		width: 50%;
	}
	.submit {
		background-color: #111111;
		color: #eeeeee;
		border-radius: 3px;
		border: none;
		padding: 1rem;
	}

	form {
		display: flex;
		flex-direction: column;
		align-items: center;
		background-color: #ffffff;
		padding: 2rem;
		border-radius: 5px;
		box-sizing: border-box;
	}

	h3 {
		display: block;
		padding: 1rem;
		width: 100%;
		padding-bottom: 1rem;
		border-bottom: solid;
		border-width: 1px;
		border-color: #ccc;
	}


	.oauth-buttons {
		display: flex;
		justify-content: space-between;
		padding-bottom: 2rem;

		button {
			display: flex;
			align-items: center;
		}
	}

	.icon {
		width: 24px;
		margin-right: 1rem;
	}
</style>
