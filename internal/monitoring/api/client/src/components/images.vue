<template>
  <div class="card">
    <header class="card-header">
      <p class="card-header-title">
        Component
      </p>
      <a href="#" class="card-header-icon" aria-label="more options">
        <span class="icon">
          <i class="fas fa-angle-down" aria-hidden="true"></i>
        </span>
      </a>
    </header>
    <div class="card-content">
      <table class="table is-striped is-hoverable">
        <thead>
          <tr>
            <th>Image</th>
            <th>CVE count</th>
            <th>Owner</th>
          </tr>
        </thead>
        <tfoot>
          <tr>
            <th>Image</th>
            <th>CVE count</th>
            <th>Owner</th>
          </tr>
        </tfoot>
        <tbody>
          <tr :key="image.id" v-for="image in list" @click="link(image.id)">
            <th>{{image.tag}}</th>
            <td>Virity found {{image.cve_count}} CVEs</td>
            <td>{{image.owner.toString()}}</td>
          </tr>
        </tbody>
      </table>
    </div> 
  </div>
</template>

<script>
export default {
  name: 'Images',
  props: {
    list: {
      type: Array,
    }
  },
  methods: {
    link(target) {
      window.open('#/image/'+target, '_self')
    }
  },
}
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped>

</style>
