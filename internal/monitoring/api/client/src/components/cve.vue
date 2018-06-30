<template>
  <div class="card is-flex is-column">
    <div class="card-content">
      <p class="title is-4 has-text-justified">
        <span class="icon" :class="severity">
          <i class="fas fa-thermometer-full" v-if="Severity == 0"></i>
          <i class="fas fa-thermometer-empty" v-else-if="Severity >= 2"></i>
          <i class="fas fa-thermometer-half" v-else></i>
        </span>
        {{Vuln}}
      </p>
      <p class="subtitle">
        {{Package | truncate(28)}}
      </p>
      <div class="content">
        <p v-if="Fix !== 'None'">Fix: {{Fix}}</p>
        <p>{{description}}</p>
      </div>
    </div>
    <footer class="card-footer">
      <p class="card-footer-item">
        <a :href="URL" target="_blank">
          <span class="is-uppercase">
            More information
          </span>
        </a>
      </p>
    </footer>
  </div>    
</template>
<script>
export default {
    name: 'Vulnerabilities',
    props: {
      Description: {
        type: String,
        default: "No description available.",
      },
      Fix: {
        type: String,
        default: "No fix available.",
      },
      Package: {
        type: String,
        default: "No package information available."
      },
      Severity: {
        type: Number,
        default: -1,
      },
      URL: {
        type: String,
        default: "https://www.google.com/search?q=broken+link",
      },
      Vuln: {
        type: String,
        default: "An Error occoured.",
      },
    },
    computed: {
      description() {
        return this.Description ? this.Description : "No description available."
      },
      severity() {
        switch (this.Severity) {
          case 0:
            return {'severity-high': true}
          case 1:
            return {'severity-medium': true}
          case 2:
          case 3:
            return {'severity-low': true}
          default: 
            return {}  
        }
      }
    }
}
</script>
<style scoped>
.is-flex.is-column {
  flex-direction: column;
  height: 100%;
  justify-content: space-between;
}

.severity-high {
  color: rgba(0, 0, 0, 1);
}

.severity-medium {
  color: rgba(0, 0, 0, 0.6);
}

.severity-low {
  color: rgba(0, 0, 0, 0.2);
}

</style>
