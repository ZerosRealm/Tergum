<script>
	export let data;
	$: type = data.name.slice(data.name.lastIndexOf('.') + 1);
</script>

<span style="background-image: url(/icons/{type}.svg)">{data.name}</span>

<style>
	span {
		padding: 0 0 0 1.5em;
		background: 0 0.1em no-repeat;
		background-size: 1em 1em;
	}
</style>