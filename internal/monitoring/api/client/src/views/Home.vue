<template>
  <div v-if="list.length === 0">
    <loader />
  </div>
  <div v-else>
    <Images :list="list"/>
  </div>
</template>

<script>
// @ is an alias to /src
import Images from '@/components/images.vue'
import Loader from '@/components/loader.vue'

export default {
  name: 'home',
  components: {
    Images,
    Loader
  },
  computed: {
    list () {
      console.log(this.$store.state.images.images)
      return this.$store.state.images.images
    },
  },
  mounted() {
    this.$store.dispatch('fetchImages')
  }
}
</script>
<style scoped>

</style>
