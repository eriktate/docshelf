import App from "./App.svelte";
import { login } from "./api";

declare global {
	interface Window {
		onSignIn: (user: any) => Promise<void>;
	}
}

window.onSignIn = async function(user: any): Promise<void> {
	const idToken = user.getAuthResponse().id_token;
	try {
		await login("", idToken); // login with no email assumes oauth
	} catch (err) {
		console.log("failed to sign in with oauth");
	}
}

const app = new App({
	target: document.body,
});

export default app;
