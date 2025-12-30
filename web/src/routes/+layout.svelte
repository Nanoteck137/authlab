<script lang="ts">
  import { Home, LogIn, LogOut, Menu } from "lucide-svelte";
  import "../app.css";
  import Link from "$lib/components/Link.svelte";
  import { browser } from "$app/environment";
  import { fade, fly } from "svelte/transition";
  import { Toaster } from "svelte-5-french-toast";
  import { setApiClient, setApiClientAuth } from "$lib";

  let { children, data } = $props();

  let apiClient = setApiClient(data.apiAddress, data.userToken);

  $effect(() => {
    if (!browser) return;
    setApiClientAuth(apiClient, data.userToken);
  });

  let showSideMenu = $state(false);

  function close() {
    showSideMenu = false;
  }

  $effect(() => {
    if (showSideMenu) {
      if (browser) document.body.style.overflow = "hidden";
    } else {
      if (browser) document.body.style.overflow = "";
    }
  });
</script>

<svelte:head>
  <title>Auth Lab</title>
</svelte:head>

<Toaster position="bottom-right" />

<header
  class="sticky top-0 z-50 w-full border-b border-border/40 bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60"
>
  <div class="container flex h-14 max-w-screen-2xl items-center gap-4">
    <button
      onclick={() => {
        showSideMenu = true;
      }}
    >
      <Menu size="20" />
    </button>

    <a class="text-2xl font-medium text-[--logo-color]" href="/">Auth Lab</a>

    <!-- <div class="flex-grow"></div>

    <div class="flex items-center gap-2">
    </div> -->
  </div>
</header>

<main class="container py-4">
  {@render children()}
</main>

{#if showSideMenu}
  <!-- svelte-ignore a11y_consider_explicit_label -->
  <button
    class="fixed inset-0 z-50 bg-black/80"
    onclick={() => {
      showSideMenu = false;
    }}
    transition:fade={{ duration: 200 }}
  ></button>

  <aside
    class={`fixed bottom-0 top-0 z-50 flex w-72 flex-col bg-sidebar text-sidebar-foreground`}
    transition:fly={{ x: -400 }}
  >
    <div class="flex h-14 items-center gap-4 border-b px-8">
      <button
        onclick={() => {
          showSideMenu = false;
        }}
      >
        <Menu size="20" />
      </button>
      <a
        class="text-2xl font-medium"
        href="/"
        onclick={() => {
          showSideMenu = false;
        }}
      >
        Dwebble
      </a>
    </div>

    <div class="flex flex-col gap-2 px-4 py-4">
      <Link title="Home" href="/" icon={Home} onClick={close} />
    </div>
    <div class="flex-grow"></div>
    <div class="flex flex-col gap-2 px-4 py-2">
      {#if data.user}
        <!-- <Link
          title={data.user.username}
          href="/account"
          icon={User}
          onClick={close}
        />

        {#if data.user.role === "super_user"}
          <Link title="Server" href="/server" icon={Server} onClick={close} />
        {/if} -->

        <form class="w-full" action="/logout" method="POST">
          <Link title="Logout" icon={LogOut} onClick={close} />
        </form>
      {:else}
        <Link title="Login" href="/login" icon={LogIn} onClick={close} />
      {/if}
    </div>
    <div class="h-4"></div>
  </aside>
{/if}
