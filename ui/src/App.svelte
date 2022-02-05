<script>
	import router from 'page'

    import socket  from './common/websocket.js';
    import { addToast }  from './common/toasts.js';

	import Home from './home/Index.svelte'
	import Repos from './repos/Repos.svelte'
	import Agents from './agents/Agents.svelte'
	import Backups from './backups/Backups.svelte'

	import Toasts from './common/Toasts.svelte'


	let currentPage = "home"

	let page
  	let params
	router('/', () => {page = Home; currentPage = "home"})
	router('/repos', () => {page = Repos; currentPage = "repos"})
	router('/agents', () => {page = Agents; currentPage = "agents"})
	router('/backups', () => {page = Backups; currentPage = "backups"})

    socket.subscribe(event => {
        if (event.data == "") {
            return
        }

        let data = JSON.parse(event.data);
        if (data.type != "error") {
            return;
        }

        addToast({
            type: "error",
            title: "Error",
            message: data.message
        });
    });

    let menuOpen = true;
    function toggleMenu() {
        menuOpen = !menuOpen;

        if (menuOpen) {
            document.querySelector("body").style.width = "100vw";
        } else {
            document.querySelector("body").style.width = "calc(100vw + 5%)";
        }
        console.log(menuOpen);
    }

  	router.start()
</script>

<style>
:global(body) {
    width: 100vw;
    overflow: hidden;
    display: inline-flex;
}

.container {
    margin: 0;
    width: 100%;
    max-width: none;
    overflow-y: auto;

    left: 0vw;
    position: relative;
    transition: left 0.5s;
}

nav {
    width: max-content;
    height: 100vh;
}

nav ul a {
    color: inherit;
    text-decoration: none;
}

nav ul {
    padding: 0;
    height: 100vh;
    list-style: none;
    background-color: #3B4252;
}

nav ul li {
    color: #fff;
    padding: 25px 0px;
    text-align: center;
    background-color: #3B4252;
    border: 2px solid transparent;
}

nav ul li:hover {
    color: #3B4252;
    cursor: pointer;
    background-color: #EDF6F9;
    border: 2px solid transparent;
}

nav ul li.active {
    color: #3B4252;
    background-color: #D8DEE9;
    border: 2px solid transparent;
}

.menu {
    left: 0%;
    position: relative;
    transition: left 0.5s;
    width: 5%
}

.menu.closed,
.container.menu-closed {
    left: -5%;
}

</style>

<svelte:body class:closed={!menuOpen} />

<div class="menu" class:closed={!menuOpen}>
    <nav>
        <ul>
            <li style="padding: 5px;" on:click={toggleMenu}>
                <svg class="bi" width="16" height="16" fill="currentColor"><use xlink:href="css/bootstrap-icons.svg#x" /></svg>
                Close
            </li>
            <a href="/"><li class:active="{currentPage == 'home'}">Home</li></a>
            <a href="/repos"><li class:active="{currentPage == 'repos'}">Repos</li></a>
            <a href="/agents"><li class:active="{currentPage == 'agents'}">Agents</li></a>
            <a href="/backups"><li class:active="{currentPage == 'backups'}">Backups</li></a>
        </ul>
    </nav>
</div>

<div class="container" class:menu-closed={!menuOpen}>
    <button class="btn btn-link p-0" on:click={toggleMenu}>
        <svg class="bi" width="32" height="32" fill="currentColor"><use xlink:href="css/bootstrap-icons.svg#list" /></svg>
    </button>
    <h1>Tergum</h1>
    <hr>
    <Toasts />
	<svelte:component this="{page}" params="{params}" />
</div>

<div id="modals"></div>