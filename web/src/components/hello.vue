<template>
  <div>
    <h2>Housekeeping</h2>
  
    <expenditures-table></expenditures-table>
  </div>
</template>

<script lang="ts">
import Vue from "vue";
import expendituresTable from '@/components/expendituresTable';

interface Hello extends Vue {
}

export default {
  name: "hello",
  data () {
    return {
    };
  },
  components: {expendituresTable},
} as Vue.ComponentOptions<Hello>;
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped>
h1,
h2 {
  font-weight: normal;
}

ul {
  list-style-type: none;
  padding: 0;
}

li {
  display: inline-block;
  margin: 0 10px;
}

a {
  color: #42b983;
}
</style>
