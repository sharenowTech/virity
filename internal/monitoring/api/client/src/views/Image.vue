<template>
  <div v-if="imageInfo === ''">
    <Loader />
  </div>
  <div v-else>
    <section class="hero is-bold">
      <div class="hero-body">
        <div class="container">
          <h1 class="title">
            {{imageInfo.tag}}
          </h1>
          <h2 class="subtitle">
            {{imageInfo.id}}
          </h2>
        </div>
      </div>
    </section>
    <section class="columns">
      <div class="column is-narrow" :key="container.ID" v-for="container in imageInfo.in_containers">
        <Container v-bind="container" />
      </div>
    </section>
    <div class="columns is-multiline">
      <div class="column is-one-quarter-desktop is-half-tablet" :key="cve.Package" v-for="cve in orderedCVEs">
        <Cve v-bind="cve" />
      </div>
    </div>
  </div>
</template>
<script>
import CVE from '@/components/cve.vue'
import Container from '@/components/container.vue'
import Loader from '@/components/loader.vue'
export default {
  name: 'cimage',
  components: {
    Cve: CVE,
    Loader,
    Container
  },
  computed: {
    imageInfo () {
      return this.$store.state.images.detail[this.$route.params.id] ? this.$store.state.images.detail[this.$route.params.id] : ''
    },
    orderedCVEs () {
      let info = this.imageInfo
      return info.vulnerability_cve.sort((a, b) => b.Severity-a.Severity);
    }
  },
  mounted() {
    this.$store.dispatch({
      type: 'fetchDetails',
      id: this.$route.params.id
    })
  }
}
</script>
<style>

</style>
