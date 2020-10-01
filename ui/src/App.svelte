<script lang="typescript">
	import { Router, Route, navigate } from "svelte-routing";
	import EditDoc from "./EditDoc.svelte";
	import DocList from "./DocList.svelte";
	import TopNav from "./TopNav.svelte";
	import Signin from "./Signin.svelte";
	import Profile from "./Profile.svelte";
	import { listDocs, getCurrentUser} from "./api.ts";

	import type { User, Doc } from "./api.ts"; //= end

	/* const email = "root@docshelf.io"; */
	/* const password = "btnslev1n0pakb16b80g"; */

	export let url: string = "";
	let user: User;
	let docs: Doc[] = [];
	let selected: string;

	async function init(): void {
		console.log("running init");
		try {
			user = await getCurrentUser();
			docs = await listDocs();
		} catch (err) {
			console.log("failed to fetch user, need to sign in!");
		}
	}

	function setUser(newUser: User): void {
		user = newUser;
	}

	function selectDoc(id: string): void {
		console.log("selecting: ", id);
		selected = id;
		navigate(`/edit/${id}`);
	}

	$: currentDoc = docs.find(doc => doc.id === selected);

	init();
</script>

<Router url={url}>
	<TopNav {user}/>
	<main>
		<Route path="/">
			<DocList {docs} {selectDoc}/>
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

	h1 {
		color: #ff3e00;
		text-transform: uppercase;
		font-size: 4em;
		font-weight: 100;
	}

	@media (min-width: 640px) {
		main {
			max-width: none;
		}
	}
</style>
