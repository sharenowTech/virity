<template>
  <div class="section">
    <div class="container">
      <div class="tabs">
        <ul>
          <router-link tag="li" to="/"><a>Home</a></router-link>
          <router-link tag="li" to="/about"><a>About</a></router-link>
        </ul>
      </div>
      <router-view/>
    </div>
  </div>
</template>

<style src="bulma"></style>
<style>

.icon.rounded {
  background-color: rgba(255, 255, 255, 0.3);
  border-radius: 50%;
}

.scroll {
  overflow-y: scroll;
  max-height: 80vh;
}

</style>
