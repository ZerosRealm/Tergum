<script>
    import Modal from '../common/Modal.svelte';
    import { callAPI }  from '../common/API.js';

    import Logs from './Logs.svelte';
    import Logging from './Logging.svelte';
    import Registration from './Registration.svelte';

</script>
<style>
    h2 {
        font-size: 1.5em;
        margin-bottom: 0.5em;
    }
</style>

<Logging />
<hr/>
<Registration />
<hr/>
<Logs />
<hr/>