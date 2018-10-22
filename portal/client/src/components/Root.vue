<template>
  <div class="container">
    <header class="container">
      <nav class="navbar level" role="navigation" aria-label="main navigation">
        <div class="navbar-brand level-left">
          <p class="level-item">
            MFAC ISUCON PORTAL
          </p>
          <p class="level-item">
            <strong>{{teamName}}</strong>
          </p>
        </div>
        <div class="navbar-menu level-right">
          <p class="level-item">
            <a href="#TODO" target="_blank">REGURATION</a>
          </p>
          <p class="level-item">
          <a class="button is-info" v-on:click="postEnqueue()">ベンチマーク実行</a>
          </p>
        </div>
      </nav>
    </header>

    <div id="error-container">
      <div class="notification is-warning" v-show="errorMessage">{{errorMessage}}</div>
    </div>

    <div id="chart-container">
      <line-chart :chart-data="datacollection" :options="chartOptions"></line-chart>
    </div>

    <div id="ranking-container" class="panel">
      <p class="panel-heading">ランキング</p>
      <div class="panel-block">
        <table class="table is-striped is-fullwidth">
          <thead>
            <tr>
              <th>順位</th>
              <th>名前</th>
              <th>最高スコア</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(ranking, index) in rankingScores" :key="index">
              <td>{{ index + 1 }}</td>
              <td>{{ ranking.team_name }}</td>
              <td>{{ ranking.max_score }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <div id="jobs-container">
      <p class="panel-heading">処理待ちのベンチマーク</p>
      <div class="panel-block">
        <table v-if="jobs.length" class="table is-striped is-fullwidth">
          <thead>
            <tr>
              <th>名前</th>
              <th>ステータス</th>
              <th>enqueued_at</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(job, index) in jobs" :key="index">
              <td>{{ job.team_name }}</td>
              <td>{{ job.status }}</td>
              <td>{{ job.enqueued_at }}</td>
            </tr>
          </tbody>
        </table>
        <div v-else>
          処理待ちのベンチーマークはありません
        </div>
      </div>
    </div>

    <div id="team-scores-container">
      <p class="panel-heading">チームスコア履歴</p>
      <div class="panel-block">
        <table class="table is-striped is-fullwidth">
          <thead>
            <tr>
              <th>通過</th>
              <th>スコア</th>
              <th>メッセージ</th>
              <th>時刻</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(teamScore, index) in teamScores" :key="index">
              <td>{{ teamScore.pass ? 'PASS' : 'FAIL' }}</td>
              <td>{{ teamScore.score }}</td>
              <td>{{ teamScore.message }}</td>
              <td>{{ teamScore.created_at }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

  </div>
</template>

<script>

import LineChart from './LineChart.js'
import axios from 'axios'

export default {
  name: 'Root',
  components: {
    LineChart
  },
  data () {
    return {
      datacollection: null,
      rankingScores: [],
      jobs: [],
      teamScores: [],
      teamName: '-',
      errorMessage: null,
      chartOptions: {
        elements: {
          line: {
            tension: 0,
            fill: false
          }
        },
        options: {
          title: 'mfac isucon'
        },
        responsive: true,
        maintainAspectRatio: false,
        scales: {
          xAxes: [{
            type: 'time',
            distribution: 'linear'
          }],
          yAxes: [
            {
              ticks: {
                beginAtZero: true,
                min: 0
              }
            }
          ]
        }
      }
    }
  },
  mounted () {
    this.fetchTeam()
    this.fetchPassScores()
    this.fetchRankingScores()
    this.fetchJobs()
    this.fetchTeamScores()
  },
  methods: {
    fetchTeam () {
      axios.get('/team', {
        withCredentials: true
      }).then(response => {
        let team = response.data
        this.teamName = team.name
      })
    },
    fetchPassScores () {
      axios.get('/pass_scores', {
        withCredentials: true
      }).then(response => {
        let scores = response.data
        let data = {
          labels: [
            '2018-09-11 10:00',
            '2018-09-11 10:30',
            '2018-09-11 11:00',
            '2018-09-11 11:30',
            '2018-09-11 12:00',
            '2018-09-11 12:30',
            '2018-09-11 13:00',
            '2018-09-11 13:30',
            '2018-09-11 14:00',
            '2018-09-11 14:30',
            '2018-09-11 15:00',
            '2018-09-11 15:30',
            '2018-09-11 16:00',
            '2018-09-11 16:30',
            '2018-09-11 17:00',
            '2018-09-11 17:30',
            '2018-09-11 18:00'
          ]
        }
        data.datasets = scores
        this.datacollection = data
      })
    },
    fetchRankingScores () {
      axios.get('/ranking_scores', {
        withCredentials: true
      }).then(response => {
        this.rankingScores = response.data
      })
    },
    fetchJobs () {
      axios.get('/jobs', {
        withCredentials: true
      }).then(response => {
        this.jobs = response.data
      })
    },
    fetchTeamScores () {
      axios.get('/team_scores', {
        withCredentials: true
      }).then(response => {
        this.teamScores = response.data
      })
    },
    postEnqueue () {
      axios.post('/enqueue', {
        withCredentials: true
      }).then(response => {
        let data = response.data
        if (data.result === 'success') {
          this.fetchJobs()
        } else {
          this.errorMessage = data.reason
        }
      })
    }
  }
}
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped>
.container {
}

#chart-container {
  height: 400px;
  margin: 40px 0;
}

#ranking-container {
  margin: 40px 0;
}

#jobs-container {
  margin: 40px 0;
}

</style>
