<script>
	import File from './File.svelte';
	import {slide} from 'svelte/transition'
	
	export let expanded = false;
	export let name;
	export let files;

	function toggle() {
		expanded = !expanded;
	}
</script>

<div class="folder" on:click={toggle}>
	<span class:expanded>{name}</span>
</div>

{#if expanded}
	<ul transition:slide={{duration:300}}>
		{#each files as file}
            {#if file.type === 'dir'}
			    <li>
					<svelte:self name={file.name} files={file.files} />
                </li>
			{/if}
		{/each}
        {#each files as file}
            {#if file.type === 'file'}
			    <li>
					<File bind:data={file}/>
                </li>
			{/if}
		{/each}
	</ul>
{/if}

<style>
	span {
		padding: 0 0 0 1.5em;
		background: url(/icons/folder.svg) 0 0.1em no-repeat;
		background-size: 1em 1em;
		font-weight: bold;
		cursor: pointer;
	}

	.folder {
		cursor: pointer;
	}

	.folder:hover {
		background: #eee;
	}

	.expanded {
		background-image: url(/icons/folder-open.svg);
	}

	ul {
		padding: 0.2em 0 0 0.5em;
		margin: 0 0 0 0.5em;
		list-style: none;
		border-left: 1px solid #eee;
	}

	li {
		padding: 0.2em 0;
	}
</style>
