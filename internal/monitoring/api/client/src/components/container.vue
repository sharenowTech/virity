<template>
  <div class="media box has-background-primary has-text-white">
    <div class="media-left">
      <span class="icon is-large is-size-4 rounded">
        <i class="fas fa-box-open"></i>
      </span>
    </div>
    <div class="media-content">
      <p class="heading">In Container</p>
      <p class="title has-text-white">{{Name}}</p>
      <div>
        <p class="is-uppercase">
          <span class="icon">
            <i class="fas fa-user"></i>
          </span>
          {{OwnerID}}
        </p>
        <p class="">
          <span class="icon">
            <i class="fas fa-map-marker-alt"></i>
          </span>
          {{Hostname}}
        </p>
        <p class="">
          <span class="icon">
            <i class="fas fa-id-card"></i>
          </span>
          {{ID | truncate(12, '...') }}
        </p>
      </div>
    </div>
  </div>
</template>
<script>
export default {
    name: 'Container',
    props: {
      Hostname: {
        type: String,
      },
      ID: {
        type: String,
      },
      ImageID: {
        type: String,
      },
      Name: {
        type: String,
      },
      OwnerID: {
        type: String,
        default: "anonymous"
      },
    },
}
</script>
<style scoped>

</style>
