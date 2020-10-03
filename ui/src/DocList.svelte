<script lang="typescript">
	import { onMount } from "svelte";
	import { link } from "svelte-routing";
	import { listDocs } from "./api";

	import type { Doc } from "./api"; // = end

	let docs: Doc[] = [];

	function friendlyTimestamp(ts?: string): string {
		if (ts) {
			const date = new Date(ts);
			return date.toDateString();
		}

		return "";
	}

	onMount(async () => {
		docs = await listDocs();
	});
</script>

<div class="pane">
	<h3>Your Documents</h3>
	<ul>
		{#if docs.length == 0}
			<div>You don't have any!</div>
		{/if}
		{#each docs as doc}
			<li>
				<a href={`/edit/${doc.id}`} use:link>
					{doc.title}
					<div>{friendlyTimestamp(doc.updatedAt)}</div>
				</a>
			</li>
		{/each}
	</ul>
</div>

<style>
	a {
		display: flex;
		justify-content: space-between;
		padding: 1rem;
		color: #222;
	}

	li {
		background-color: #fff;
		border-radius: 2px;
		border-style: solid;
		border-width: 1px;
		border-color: #ccc;
	}

	li:hover {
		background-color: #ccc;
	}

	.pane {
		padding: 1rem;
	}

</style>
