<template>
  <div v-if="imageInfo === ''">
    <div class="section is-medium container has-text-centered">
      <span class="icon">
        <i class="fas fa-spinner fa-pulse"></i>
      </span>
    </div>
  </div>
  <div v-else>
    <h1 class="title">{{imageInfo.tag}}</h1>
    <h2 class="subtitle">{{imageInfo.id}}</h2>
    
    <p :key="cve.Package" v-for="cve in imageInfo.vulnerability_cve">
      {{cve.Vuln}}
    </p>
    <p :key="owner" v-for="owner in imageInfo.owner">
      {{owner}}
    </p>
    <p :key="container.ID" v-for="container in imageInfo.in_containers">
      {{container.ID}}
    </p>

  </div>
</template>
<script>
export default {
  name: 'cimage',
  computed: {
    imageInfo () {
      return this.$store.state.images.detail[this.$route.params.id] ? this.$store.state.images.detail[this.$route.params.id] : ''
    },
  },
  mounted() {
    this.$store.dispatch({
      type: 'fetchDetails',
      id: this.$route.params.id
    })
  }
}
</script>
<style scoped>

.icon {
  font-size: 5rem;
}

</style>
