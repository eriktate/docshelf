<script lang="typescript">
	import type { Doc } from "./api.ts"; // = end
	export let docs: Doc[] = [];
	export let selectDoc: (id: string) => void;

	function friendlyTimestamp(ts): string {
		const date = new Date(ts);
		return date.toDateString();
	}
</script>

<h3>Your Documents</h3>
<ul>
	{#each docs as doc}
		<li><a on:click={selectDoc(doc.id)}>{doc.title}</a> - {friendlyTimestamp(doc.updatedAt)}</li>
	{/each}
</ul>
