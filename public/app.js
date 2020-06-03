/* global jQuery, Vue, PNotify, io */
// pnotify
PNotify.prototype.options.styling = 'fontawesome'
function showNotify(o) {
  var opts = {
    animate_speed: 'fast',
    buttons: {
      closer: true,
      sticker: false
    }
  }
  if (arguments.length > 1) {
    if (arguments.length === 2) {
      o = {
        type: arguments[0],
        text: arguments[1]
      }
    } else {
      o = {
        type: arguments[0],
        title: arguments[1],
        text: arguments[2]
      }
    }
  } else {
    if (typeof (o) === 'string') {
      o = {
        type: 'info',
        text: o
      }
    }
  }
  opts = Object.assign(opts, o)
  if (!opts.hide) {
    opts.animation = 'none'
  }
  var notice = new PNotify(opts)
  if (o.clickToClose !== false) {
    notice.get().click(function () {
      notice.remove()
    })
  }
  return notice
}
// eslint-disable-next-line no-unused-vars
var app;
!(function () {
  var access_token;
  app = new Vue({
    el: '#app',
    data: {
      loggedin: false,
      user: {},
      currentTab: null,
      tabs: [{
        name: 'Home',
        component: 'home'
      }, {
        name: 'Quản lý phiên',
        component: 'sessions'
      }],
      sessions: [],
      current_token_id: ''
    },
    mounted: function () {
      var self = this
      self.currentTab = ''
    },
    watch: {
      currentTab: function (newVal) {
        // array.some ? no, es5 :(
        var self = this
        if (!this.loggedin && !['login', 'register'].includes(newVal)) {
          self.currentTab = 'login'
          return
        }
        if (newVal == '') {
          self.currentTab = 'home'
        } else if (newVal === 'sessions') {
          self.getSessions()
        }
      },
      loggedin: function (val) {
        if (val === false) {
          this.currentTab = 'login'
        }
      }
    },
    methods: {
      clearTokens() {
        fetch('/api/tokens', {
          method: 'DELETE',
          headers: {
            'X-Token': access_token
          },
          body: 'all'
        }).then(r => r.json()).then(r => {
          if (r.error) {
            showNotify('error', r.error)
            return
          }
          this.sessions = this.sessions.filter(x => x.id == this.current_token_id)
          showNotify('success', 'Đã đăng xuất khỏi các thiết bị khác!')
        }).catch(e => {
          showNotify('error', e + '')
        })
      },
      delToken(tid, i) {
        fetch('/api/tokens', {
          method: 'DELETE',
          headers: {
            'X-Token': access_token
          },
          body: tid
        }).then(r => r.json()).then(r => {
          if (r.error) {
            showNotify('error', r.error)
            return
          }
          this.$delete(this.sessions, i)
          if (tid == this.current_token_id) {
            showNotify('success', 'Đã đăng xuất!')
            access_token = ''
            this.current_token_id = ''
            this.loggedin = false
          }
        }).catch(e => {
          showNotify('error', e + '')
        })
      },
      getSessions() {
        this.sessions = []
        var notice = showNotify({
          type: 'info',
          text: 'Đang tải sessions..',
          hide: false,
          clickToClose: false
        })
        fetch('/api/tokens', {
          headers: {
            'X-Token': access_token
          }
        }).then(r => r.json()).then(r => {
          notice.remove()
          if (r.error) {
            showNotify('error', r.error)
            return
          }
          this.sessions = r
        }).catch(e => {
          notice.remove()
          showNotify('error', e + '')
        })
      },
      getAccessTokenId(at) {
        try {
          return JSON.parse(atob(at.split('.')[1]))['id']
        } catch (e) {
          console.error(e)
        }
        return ''
      },
      logout() {
        fetch('/api/logout', {
          method: 'POST',
          headers: {
            'X-Token': access_token
          }
        }).then(r => r.json()).then(r => {
          if (r.error) {
            showNotify('error', r.error)
            return
          }
          showNotify('success', 'Đã đăng xuất!')
          access_token = ''
          this.current_token_id = ''
          this.loggedin = false
        }).catch(e => {
          showNotify('error', e + '')
        })
      },
      login: function ($e) {
        var form = $e.target
        var username = form.username.value
        var password = form.password.value
        var notice = showNotify({
          type: 'info',
          text: 'Đang đăng nhập..',
          hide: false,
          clickToClose: false
        })
        fetch('/login', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/x-www-form-urlencoded'
          },
          body: `username=${encodeURIComponent(username)}&password=${encodeURIComponent(password)}`
        }).then(r => r.json()).then(r => {
          notice.remove()
          if (r.error) {
            showNotify('error', r.error)
            return
          }
          showNotify('success', 'Đăng nhập thành công!')
          access_token = r.access_token
          this.current_token_id = this.getAccessTokenId(access_token)
          fetch('/api/me', {
            headers: {
              'X-Token': access_token
            }
          }).then(r => r.json()).then(r => {
            this.user = r
            this.loggedin = true
            this.currentTab = 'home'
          }).catch(e => {
            showNotify('error', e + '')
          })
        }).catch(e => {
          notice.remove()
          showNotify('error', e + '')
        })
      },
      register($e) {
        var form = $e.target
        var username = form.username.value
        var password = form.password.value
        var notice = showNotify({
          type: 'info',
          text: 'Đang đăng ký..',
          hide: false,
          clickToClose: false
        })
        fetch('/register', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/x-www-form-urlencoded'
          },
          body: `username=${encodeURIComponent(username)}&password=${encodeURIComponent(password)}`
        }).then(r => r.json()).then(r => {
          notice.remove()
          if (r.error) {
            showNotify('error', r.error)
            return
          }
          showNotify('success', 'Đăng ký thành công!')
        }).catch(e => {
          notice.remove()
          showNotify('error', e + '')
        })
      },
    }
  })
})()