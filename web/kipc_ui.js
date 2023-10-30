let proto = document.location.protocol === 'https:' ? 'wss://' : 'ws://'
let ws = new WebSocket(proto + document.location.host + "/ws")
ws.onopen = function() { app.log.push({w:0, m:'connecting...'}) }
ws.onclose = function() { app.log.push({w:0, m:'connection closed.'}) }
ws.onmessage = function(e) {
  let [txid, err, msg] = JSON.parse(e.data), txn = app.log[txid];
  if (txn.w) Vue.set(txn, 'w', 0); // clear waiting flag
  if (err) Vue.set(txn, 'e', msg); else Vue.set(txn, 'o', msg); }

var app = new Vue({
  el: '#app',
  data: {
    n:1, history:[], histCursor:-1, waitsym:0,
    log: [], draft:'', input: '{*+ x (+\\|:)\\ 1 1} 16' },
  methods: {

    send(k) {
      // send messages starting with \! as async. for others, expect immediate response
      let n = this.n, entry = {n:n, i:k};
      if (k.startsWith('\\! ')) { entry.w = 1; ws.send(JSON.stringify([n, 1, k.slice(3)])) }
      else ws.send(JSON.stringify([n, 0, k]));
      this.log.push(entry); this.n+=1 },

    histMax() { return this.history.length-1 },

    enter() {
      if (this.input.length) this.send(this.input);
      if (this.histCursor < 0 || this.input !== this.history[this.histMax()]) {
        this.history.push(this.input); }
      this.histCursor = this.history.length; this.draft=''; this.input = ''; },

    /// move up in history
    up() { if (this.histCursor > 0) {
      if (this.histCursor === this.history.length) this.draft = this.input;
      this.histCursor -= 1; this.input = this.history[this.histCursor] }},

    /// move down in history
    dn() {
      if (this.histCursor === this.histMax()) { this.input = this.draft; }
      this.histCursor += 1;
      if (this.histCursor < this.history.length) this.input = this.history[this.histCursor];
      else { this.histCursor = this.history.length;  }},

    // handle special keys
    keyup(e) {
      switch(e.key) {
        case "Enter": this.enter(); break;
        case "ArrowUp": this.up(); break;
        case "ArrowDown": this.dn(); break;
        default: return false
      }
    }
  }
});

setInterval(function(){ app.$data.waitsym = (app.$data.waitsym+1) & 7; }, 250)
app.$refs.prompt.focus()
