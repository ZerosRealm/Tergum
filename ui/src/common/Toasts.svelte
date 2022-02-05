<script>
    import { subscribe } from 'svelte/internal';
    import { toasts } from './toasts.js';

    function close(event) {
        let toast = event.target.parentNode.parentNode;
        toast.parentNode.removeChild(toast);
    }

    const onload = el => {
        setTimeout(() => {
            if (el.parentNode != null) {
                el.parentNode.removeChild(el);
            }
        }, 5000);
    }

    const capitalize = (s) => {
        if (typeof s !== 'string') return ''
        return s.charAt(0).toUpperCase() + s.slice(1)
    }
</script>
<style>
    .bg-info,
    .bg-success,
    .bg-danger {
        color: white;
    }
</style>

<div class="toast-container position-fixed bottom-0 end-0 p-3" style="z-index: 11">
    {#each $toasts as toast}
        <div class="toast showing" use:onload role="alert" aria-live="assertive" aria-atomic="true">
            <div class="toast-header" class:bg-danger={toast.type === 'error'} class:bg-success={toast.type === 'success'} class:bg-info={toast.type === 'info'}>
                <!-- svelte-ignore missing-declaration -->
                <strong class="me-auto">{capitalize(toast.title)}</strong>
                <button type="button" class="btn-close" data-bs-dismiss="toast" aria-label="Close" on:click={close}></button>
            </div>
            <div class="toast-body">
                <!-- svelte-ignore missing-declaration -->
                {capitalize(toast.message)}
            </div>
        </div>
    {/each}
</div>