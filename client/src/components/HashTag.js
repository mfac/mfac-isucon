export default {
  name: 'HashTag',
  props: [ 'body' ],
  template: '<component v-bind:is="linked"></component>',
  computed: {
    linked () {
      const template = this.hashtag2link(this.body)
      return {
        template: template,
        props: this.$options.props
      }
    }
  },
  methods: {
    hashtag2link (str) {
      if (str) {
        return '<div>' + str.replace(/#(\S*)/g, '<router-link to="/tag/$1">#$1</router-link>') + '</div>'
      }
    }
  }
}
