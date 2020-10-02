<script lang="typescript">
	import { onMount } from "svelte";
	import { link } from "svelte-routing";
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

<div class="pane">
	<h3>Your Documents</h3>
	<ul>
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
