<script lang="typescript">
	import { Router, Route } from "svelte-routing";
	import EditDoc from "./EditDoc.svelte";
	import DocList from "./DocList.svelte";
	import TopNav from "./TopNav.svelte";
	import Signin from "./Signin.svelte";
	import Profile from "./Profile.svelte";
	import { getCurrentUser} from "./api";

	import type { User } from "./api"; //= end

	export let url: string = "";
	let user: User;

	async function init(): Promise<void> {
		try {
			user = await getCurrentUser();
		} catch (err) {
			console.log("failed to fetch user, need to sign in!");
		}
	}

	init();
</script>

<Router url={url}>
	<TopNav {user}/>
	<main>
		<Route path="/">
			<DocList />
		</Route>
		<Route path="/edit/:id" component={EditDoc} />
		<Route path="/create" component={EditDoc} />
		<Route path="/signin">
			<Signin success={init}/>
		</Route>
		<Route path="/profile">
			<Profile {user} />
		</Route>
	</main>
</Router>

<style>
	main {
		max-width: 240px;
		height: calc(100% - 4rem);
		padding: 2rem;
		box-sizing: border-box;
	}

	@media (min-width: 640px) {
		main {
			max-width: none;
		}
	}
</style>
