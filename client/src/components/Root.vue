<template>
  <div id="root-container" class="container">
    <loading :active.sync="isLoading"
    :can-cancel="true"
    :on-cancel="whenCancelled"
    :is-full-page="fullPage"></loading>

    <div class="columns">
      <div id="info-container" class="column">
        <h1><router-link to="/">geomemo</router-link></h1>
        <p>
          地図にメモするアプリ
        </p>
      </div>
      <div id="map-container" class="column">
        <div id="map"></div>
      </div>
      <div id="timeline-container" class="column">
        <div v-if="memos.length">
          <div class="card" v-for="(memo,index) in memos" :key="index">
            <div class="card-image">
              <figure class="image is-4by3">
                <img src="http://placehold.jp/400x300.png" alt="Placeholder image">
              </figure>
            </div>
            <div class="card-content">
              <div class="content">
                <hash-tag :body="memo.body" />
                <div>
                  <div v-for="emj in memo.emojis" :key="emj">
                    {{ emj.emoji }}:{{ emj.count }}
                  </div>
                  <button class="button" @click="showPicker(index)">add reaction</button>
                </div>
                ({{ memo.lat }}, {{ memo.lng }})
                <time>{{ memo.created_at }}</time>
                <router-link :to="{ name: 'Around', params: { memo_id: memo.id }}">周辺のメモを見る</router-link>
              </div>
            </div>
          </div>
          <div class="control">
            <button class="button" @click="moreMemo()" v-show="hasNext">もっと見る</button>
          </div>
        </div>
        <div v-else>
          メモがありません
        </div>
      </div>
    </div>

    <div class="modal" v-bind:class="{ 'is-active': availablePicker }" >
      <div class="modal-background" @click="hidePicker()"></div>
      <div class="modal-content">
        <picker native="TRUE" @select="addEmoji" />
      </div>
      <button class="modal-close is-large" aria-label="close" @click="hidePicker()"></button>
    </div>

    <div class="modal" v-bind:class="{ 'is-active': availableMemoForm }" >
      <div class="modal-background" @click="hideMemoForm()"></div>
      <div class="modal-card">
        <header class="modal-card-head">
           <p class="modal-card-title">メモする</p>
           <button class="delete" aria-label="close" @click="hideMemoForm()"></button>
        </header>
        <section class="modal-card-body">
          <div class="control">
            <div class="field is-grouped">
              <label class="label">緯度</label>
              <div class="control">
                <input class="input" type="text" v-model="newMemo.lat" name="lat" />
              </div>
              <label class="label">経度</label>
              <div class="control">
                <input class="input" type="text" v-model="newMemo.lng" name="lng" />
              </div>
            </div>
            <div class="field">
              <textarea class="textarea" name="body" v-model="newMemo.body"></textarea>
            </div>
          </div>
        </section>
        <footer class="modal-card-foot">
          <button class="button is-success" @click="postMemo">メモする</button>
          <button class="button" @click="hideMemoForm()">キャンセル</button>
        </footer>
      </div>
    </div>

  </div>
</template>

<script>
import leaflet from 'leaflet'
import { Picker } from 'emoji-mart-vue'
import axios from 'axios'
import Loading from 'vue-loading-overlay'
import 'vue-loading-overlay/dist/vue-loading.min.css'
import HashTag from './HashTag.js'

export default {
  name: 'Root',
  components: {
    Picker,
    Loading,
    HashTag
  },
  data () {
    return {
      memos: [],
      map: null,
      marker: null,
      availablePicker: false,
      availableMemoForm: false,
      selectedMemoIndex: null,
      newMemo: {},
      page: 1,
      hasNext: null,
      isLoading: false
    }
  },
  watch: {
    '$route' (to, from) {
      this.fetchMemos()
    }
  },
  mounted () {
    this.fetchMemos()
    this.setupMap()
  },
  methods: {
    setupMap () {
      let map = leaflet.map('map').setView([51.505, -0.09], 13)

      leaflet.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
        attribution: '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors'
      }).addTo(map)

      leaflet.marker([51.505, -0.09]).addTo(map)
        .bindPopup('クリックした場所にメモできます')
        .openPopup()

      map.on('click', this.showMemoForm)

      this.map = map
    },
    fetchMemos (page = 1) {
      let url = this.$route.name === 'Tag' ? this.$route.path
        : this.$route.name === 'Around' ? this.$route.path
          : '/memo'

      this.isLoading = true
      axios.get(url, {
        params: {
          page: page
        }
      }).then(response => {
        if (page === 1) {
          this.memos = response.data.memos
          this.page = 1
          this.hasNext = response.data.has_next
        } else {
          this.memos.concat(response.data.memos)
          this.page = page
          this.hasNext = response.data.has_next
        }
        this.isLoading = false
      })
    },
    addEmoji (emoji) {
      let memo = this.memos[this.selectedMemoIndex]

      let url = '/memo/' + memo.id + '/emoji/' + emoji.id

      this.isLoading = true
      axios.post(url, {
      }).then(response => {
        this.fetchMemos(this.page)
        this.isLoading = false
      })
    },
    showPicker (memoIndex) {
      this.selectedMemoIndex = memoIndex
      this.availablePicker = true
    },
    hidePicker () {
      this.availablePicker = false
    },
    showMemoForm (e) {
      this.newMemo.lat = e.latlng.lat
      this.newMemo.lng = e.latlng.lng

      this.availableMemoForm = true
    },
    hideMemoForm () {
      this.availableMemoForm = false
    },
    postMemo (e) {
      this.isLoading = true
      axios.post('/memo', {
        lat: this.newMemo.lat,
        lng: this.newMemo.lng,
        body: this.newMemo.body
      }).then(response => {
        this.memos.unshift(response.data)
        this.newMemo = {}
        this.hideMemoForm()
        this.isLoading = false
      })
    },
    moreMemo () {
      let page = this.page + 1
      this.fetchMemos(page)
    }
  }
}
</script>

<style>
@import "../../node_modules/leaflet/dist/leaflet.css";

/* map pane が 700 までっぽいのでそれより上で */
/* https://leafletjs.com/reference-1.3.4.html#map-mappane */
.modal {
  z-index: 2000
}

.modal-background {
  background-color: rgba(10, 10,10, 0.2);
}

#root-container {
  position: relative;
}

#info-container {
  max-width: 50px;
  -webkit-writing-mode: vertical-rl;
  -ms-writing-mode: tb-rl;
  writing-mode: vertical-rl;
}

#map-container {
}

#map {
  height: 100vh;
}

#timeline-container {
  height: 100vh;
  max-width: 350px;
  overflow: scroll;
}

#timeline-container .card {
  margin: 10px 0;
}

</style>
