<template>
  <canvas ref="canvas">
  </canvas>
</template>

<script lang="ts">
import Vue from "vue";
import Chart from 'chart.js';

export default {
  name: "stats-chart",
  props: ['stats'],
  data () {
    return {
      chart: null,
    };
  },
  mounted () {
    let self = this;
    this.$watch('stats', () => {
      self.loadChart();
    });

    self.loadChart();
  },

  methods: {
    loadChart() {
      if(this.chart) {
        this.chart.destroy();
      }
      
      this.chart = new Chart(this.$refs.canvas.getContext('2d'),{
          type: 'pie',
          data: {
            labels: this.stats.map((s) => s.name || 'geen'),
            datasets: [
              {
                data: this.stats.map((s) => s.total),
                backgroundColor: this.nColors(this.stats.length),
              }
            ],
          },
          options: {}
      });
    },

    nColors(n : number) : Array<string> {
      let colors: Array<string> = [];

      for(let i = 0; i < n; i++) {
        let h = 360/n * i;
        let s = 50;
        let l = 50;

        colors.push('hsl(' + h.toString() + ',' + s.toString() + '%,' + l.toString() + '%)');
      }

      return colors;
    },
  },
};
</script>

<style lang="scss" scoped>

</style>
