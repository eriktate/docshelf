<script lang="typescript">
	import type { User } from "./api"; // = end
	export let user: User;
</script>

<div>
	ID: {user?.id}
</div>
<div>
	Name: {user?.name}
</div>
<div>
	Email: {user?.email}
</div>
