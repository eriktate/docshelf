<script lang="typescript">
	import { onMount } from "svelte";
	import { Link } from "svelte-routing";
	import { listDocs } from "./api.ts";

	import type { Doc } from "./api.ts"; // = end

	let docs: Doc[] = [];

	function friendlyTimestamp(ts): string {
		const date = new Date(ts);
		return date.toDateString();
	}

	onMount(async () => {
		docs = await listDocs();
	});
</script>

<h3>Your Documents</h3>
<ul>
	{#each docs as doc}
		<li><Link to={`/edit/${doc.id}`}>{doc.title}</Link> - {friendlyTimestamp(doc.updatedAt)}</li>
	{/each}
</ul>
