<template>
  <div v-if="list === undefined">
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
import { mapGetters } from "vuex";

export default {
  name: "home",
  components: {
    Images,
    Loader
  },
  /*filters: {
    sort: function (list) {
      return list.sort((a, b) => parseFloat(a.cve_count) - parseFloat(b.cve_count));
    }
  },*/
  computed: {
    ...mapGetters({
      list: 'getImageListSorted'
    })
  },
  methods: {
    loadData() {
      //console.log("load data");
      this.$store.dispatch("fetchImageList");
    }
  },
  mounted() {
    this.loadData();
    setInterval(() => this.loadData(), 10000);
  }
};
</script>
<style scoped>
</style>
