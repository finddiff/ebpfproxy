import { createRouter, createWebHashHistory } from 'vue-router'
import OverviewTab from './components/OverviewTab.vue'
import DhcpTab from './components/DhcpTab.vue'
import SensorsTab from './components/SensorsTab.vue'
import ProxyTab from './components/ProxyTab.vue'
import RulesTab from './components/RulesTab.vue'
import ConnectionsTab from './components/ConnectionsTab.vue'
import ConfigEditor from './components/ConfigEditor.vue'
import LogsTab from './components/LogsTab.vue'
import DnsTab from './components/DnsTab.vue'

const routes = [
  { path: '/overview',    name: 'overview',    component: OverviewTab },
  { path: '/dhcp',        name: 'dhcp',        component: DhcpTab },
  { path: '/sensors',     name: 'sensors',     component: SensorsTab },
  { path: '/proxy',       name: 'proxy',       component: ProxyTab },
  { path: '/rules',       name: 'rules',       component: RulesTab },
  { path: '/dns',         name: 'dns',         component: DnsTab },
  { path: '/connections', name: 'connections', component: ConnectionsTab },
  { path: '/config',      name: 'config',      component: ConfigEditor },
  { path: '/logs',        name: 'logs',        component: LogsTab },
  { path: '/:pathMatch(.*)*', redirect: '/overview' },
]

export default createRouter({
  history: createWebHashHistory(),
  routes,
})
