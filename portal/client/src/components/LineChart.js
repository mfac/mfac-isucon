import { Line, mixins } from 'vue-chartjs'

export default {
  name: 'LineChart',
  extends: Line,
  mixins: [mixins.reactiveData],
  props: ['chartData', 'options'],
  mounted () {
    this.renderChart(this.chartData, this.options)
  }
}
