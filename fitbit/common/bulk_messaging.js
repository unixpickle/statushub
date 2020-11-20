import * as messaging from 'messaging';

const CHUNK_SIZE = 512;

class BulkMessaging {
  constructor() {
    this.onOpen = () => null;
    this.onMessage = (msg) => null;

    this._currentId = null;
    this._buffer = '';

    messaging.peerSocket.addEventListener('open', () => this.onOpen());
    messaging.peerSocket.addEventListener('message', (evt) => {
      const data = evt.data;
      if (data['id'] !== this._currentId) {
        this._currentId = data['id'];
        this._buffer = data['data'];
      } else {
        this._buffer += data['data'];
      }
      if (data['done']) {
        this._flush();
      }
    });
  }

  isOpen() {
    return messaging.peerSocket.readyState === messaging.peerSocket.OPEN;
  }

  send(obj) {
    if (!this.isOpen()) {
      return false;
    }
    const data = JSON.stringify(obj);
    const outId = new Date().getTime() / 1000 + Math.random();
    for (let i = 0; i < data.length; i += CHUNK_SIZE) {
      if (i + CHUNK_SIZE >= data.length) {
        messaging.peerSocket.send({
          'id': outId,
          'done': true,
          'data': data.slice(i),
        });
      } else {
        messaging.peerSocket.send({
          'id': outId,
          'done': false,
          'data': data.slice(i, i + CHUNK_SIZE),
        });
      }
    }
    return true;
  }

  _flush() {
    const obj = JSON.parse(this._buffer);
    this._buffer = '';
    this._currentId = null;
    this.onMessage({ 'data': obj });
  }
}

export let peerSocket = new BulkMessaging();
