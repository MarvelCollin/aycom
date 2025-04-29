<script lang="ts">
  import SidebarNav from "../navigation/SidebarNav.svelte";
  import type { IUserProfile } from "../../interfaces/ISocialMedia";

  export let user: IUserProfile;
</script>

<div class="fixed h-screen w-[275px] flex flex-col overflow-y-auto">
  <SidebarNav {user} />
</div> 