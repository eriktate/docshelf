<script lang="ts">
	import type { User } from "./api"; //= end
	import { link } from "svelte-routing";
	import { logout } from "./api";

	export let user: User;
</script>

<nav>
	<ul>
		<li class="banner"><a href="/" use:link>Docshelf</a></li>
		<li><a href="/create" use:link>Create New</a></li>
	</ul>

	<ul class="menu">
		<li><input type="text" placeholder="Search..." /></li>
		<li><a href="/" use:link>Dashboard</a></li>
		{#if !user}
			<li><a href="/signin" use:link>Sign In</a></li>
		{:else}
			<li><a href="/profile" use:link>Profile</a></li>
		{/if}
		<li><a href="/about">About</a></li>
		{#if user}
			<li><a href="/" on:click={logout}>Log Out</a></li>
		{/if}
	</ul>
</nav>

<style>
	nav {
		display: flex;
		background-color: #111111;
		color: #eeeeee;
		height: 4rem;
		font-weight: 700;
		box-shadow: 1px 5px 5px 0 #777777;
	}

	ul {
		display: flex;
		height: 100%;
	}

	li {
		display: flex;
		align-items: center;
		height: 100%;
		line-height: 1rem;
		margin: auto 0;
		vertical-align: center;
	}

	input {
		background-color: #333;
		border-color: #333;
		color: #fff;
	}

	a {
		display: flex;
		padding: 0 1.5rem;
		align-items: center;
		height: 100%;
		color: #EEEEEE;
	}

	a:hover {
		background-color: #555555;
	}

	.banner {
		font-weight: 900;
	}

	.menu {
		margin-left: auto;
	}
</style>
