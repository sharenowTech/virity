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
import Images from "@/components/images.vue";
import Loader from "@/components/loader.vue";
import { mapState } from "vuex";

export default {
  name: "home",
  components: {
    Images,
    Loader
  },
  computed: {
    ...mapState({
      list: state => state.images.list
    })
  },
  methods: {
    loadData() {
      this.$store.dispatch("fetchImageList");
    }
  },
  mounted() {
    this.loadData();
    setInterval(() => this.loadData().bind(this), 5000);
  }
};
</script>
<style scoped>
</style>
