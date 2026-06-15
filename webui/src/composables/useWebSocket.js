import { ref, onMounted, onUnmounted } from 'vue'

export function useWebSocket(getWsUrl) {
  const connected = ref(false)
  const overviewData = ref({
    cpu_percent: 0, mem_used: 0, mem_total: 0, mem_percent: 0,
    connections: 0, udp_sessions: 0, upload_rate: 0, download_rate: 0,
    upload_total: 0, download_total: 0, net_sent: 0, net_recv: 0,
    uptime: 0, load_1: 0, load_5: 0, load_15: 0, timestamp: 0, traffic_samples: [],
  })
  const logEntry = ref(null)

  let ws = null
  let reconnectTimer = null

  function connect() {
    const url = getWsUrl('/ws')

    ws = new WebSocket(url)

    ws.onopen = () => { connected.value = true }

    ws.onclose = () => {
      connected.value = false
      scheduleReconnect()
    }

    ws.onerror = () => { ws?.close() }

    ws.onmessage = (event) => {
      try {
        const msg = JSON.parse(event.data)
        switch (msg.type) {
          case 'overview': overviewData.value = msg.data; break
          case 'log_entry': logEntry.value = msg.data; break
        }
      } catch (e) {}
    }
  }

  function scheduleReconnect() {
    if (reconnectTimer) return
    reconnectTimer = setTimeout(() => { reconnectTimer = null; connect() }, 3000)
  }

  onMounted(() => { connect() })

  onUnmounted(() => {
    if (reconnectTimer) clearTimeout(reconnectTimer)
    if (ws) ws.close()
  })

  return { connected, overviewData, logEntry }
}
