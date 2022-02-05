<script>
	import { createEventDispatcher, onDestroy, onMount } from 'svelte';

	const dispatch = createEventDispatcher();
	const close = () => dispatch('close');

	export let fit;

	let modal;
	let comp;

	const handle_keydown = e => {
		// if (e.key === 'Escape') {
		// 	close();
		// }

		if (e.key === 'Tab') {
			// trap focus
			const nodes = modal.querySelectorAll('*');
			const tabbable = Array.from(nodes).filter(n => n.tabIndex >= 0);

			let index = tabbable.indexOf(document.activeElement);
			if (index === -1 && e.shiftKey) index = 0;

			index += tabbable.length + (e.shiftKey ? -1 : 1);
			index %= tabbable.length;

			tabbable[index].focus();
			e.preventDefault();
		}
	};

	const previously_focused = typeof document !== 'undefined' && document.activeElement;

	if (previously_focused) {
		onDestroy(() => {
			previously_focused.focus();
		});
	}

	onMount(async () => {
		if (comp) {
			document.querySelector("#modals").appendChild(comp);
		}
	})
</script>

<svelte:window on:keydown={handle_keydown}/>
<div bind:this={comp}>

	<div class="modal-background" on:click={close}></div>

	<div class="showModal" role="dialog" aria-modal="true" class:fit={fit} bind:this={modal}>
		<slot name="header"></slot>
		<hr>
		<slot></slot>
		<hr>

		<slot name="buttons">
			<!-- svelte-ignore a11y-autofocus -->
			<button type="button" class="btn btn-secondary" autofocus on:click={close}>Close</button>
		</slot>
	</div>
</div>

<style>
	hr {
		height: 2px !important;
		opacity: .5;
		background-color: #2e3440;
	}
	.modal-background {
		position: fixed;
		top: 0;
		left: 0;
		width: 100%;
		height: 100%;
		background: rgba(0,0,0,0.3);
	}

	.showModal {
		color: #000;

		position: absolute;
		left: 50%;
		top: 50%;
		width: calc(100vw - 4em);
		max-width: 32em;
		max-height: calc(100vh - 4em);
		overflow: auto;
		transform: translate(-50%,-50%);
		padding: 1em;
		border-radius: 0.2em;
		background: #fff;
	}

	.fit {
		width: auto;
		max-width: fit-content;
	}
</style>