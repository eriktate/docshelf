<script lang="typescript">
	import Quill from "quill";
	import { onMount } from "svelte";
	import { createDoc, getDoc } from "./api.ts";
	import type { Doc } from "./api.ts"; // = end

	export let id = "";

	let doc: Doc = {
		title: "",
		path: "",
		content: "",
	};

	let quill;

	$: path = doc.title.toLowerCase().split(" ").join("-");

	async function init(): void {
		if (id) {
			doc = await getDoc(id);
			const content = JSON.parse(doc.content);
			quill.setContents(content);
		}
	}

	async function handleSubmit(): Promise<void> {
		const content = JSON.stringify(quill.getContents());

		doc.path = path;
		doc.content = content;
		doc.id = await createDoc(doc);
	}

	onMount(async () => {
		quill = new Quill("#editor", { theme: "snow" });
		init();
	});

	$: headerText = doc.id ? "editing document" : "creating document";
	$: buttonText = doc.id ? "Update Document" : "Create Document";
</script>

<div class="pane">
	<h3>{headerText}</h3>
	<form on:submit|preventDefault={handleSubmit}>
		<div class="controls">
			<fieldset>
				<label for="title">Title</label>
				<input name="title" placeholder="Title" type="text" bind:value={doc.title} />
				{#if path}
					<span>
						{path}
					</span>
				{/if}
			</fieldset>
			<button class="submit" type="submit">{buttonText}</button>
		</div>
		<div id="editor" />
	</form>
</div>

<style>
	#editor {
		min-height: 20rem;
	}

	.pane {
		background-color: #ffffff;
		border-radius: 5px;
		box-shadow: 1px 5px 5px 0 #777777;
		box-sizing: border-box;
	}

	.controls {
		display: flex;
		flex-direction: row;
		justify-content: space-between;
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

</style>
