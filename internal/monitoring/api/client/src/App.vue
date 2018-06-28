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

<style lang="sass" src="bulma">

</style>
