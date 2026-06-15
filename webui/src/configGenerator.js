export function generateConfig(state) {
  let text = ''

  // Global section
  text += sectionStart('global')
  for (const [key, value] of Object.entries(state.global)) {
    text += `    ${key}: ${formatValue(value, key)}\n`
  }
  text += sectionEnd()

  // Subscription section
  if (state.subscription.length > 0) {
    text += sectionStart('subscription')
    for (const item of state.subscription) {
      text += `    '${item}'\n`
    }
    text += sectionEnd()
  }

  // Node section
  if (state.node.length > 0) {
    text += sectionStart('node')
    for (const item of state.node) {
      // Check if item has tag prefix: "tag:link"
      const tagIdx = item.indexOf(':')
      if (tagIdx > 0 && tagIdx < item.indexOf('://')) {
        const tag = item.substring(0, tagIdx)
        const link = item.substring(tagIdx + 1)
        text += `    ${tag}: '${link}'\n`
      } else {
        text += `    '${item}'\n`
      }
    }
    text += sectionEnd()
  }

  // Group section
  if (state.group.length > 0) {
    text += 'group {\n'
    for (const group of state.group) {
      text += `    ${group.name} {\n`
      if (group.policy) {
        text += `        policy: ${group.policy}\n`
      }
      for (const filter of (group.filters || [])) {
        text += `        filter: ${filter}\n`
      }
      for (const extra of ['tcp_check_url', 'tcp_check_http_method', 'udp_check_dns', 'check_interval', 'check_tolerance']) {
        if (group[extra]) {
          text += `        ${extra}: '${group[extra]}'\n`
        }
      }
      text += '    }\n'
    }
    text += '}\n\n'
  }

  // DNS section
  text += sectionStart('dns')
  for (const [key, value] of Object.entries(state.dns)) {
    if (['upstream', 'hosts', 'fixed_domain_ttl', 'request_rules', 'response_rules', 'request_fallback', 'response_fallback'].includes(key)) continue
    text += `    ${key}: ${formatValue(value, key)}\n`
  }

  // DNS upstream
  if (state.dns.upstream.length > 0) {
    text += '    upstream {\n'
    for (const item of state.dns.upstream) {
      text += `        ${item.key}: '${item.value}'\n`
    }
    text += '    }\n'
  }

  // DNS hosts
  if (state.dns.hosts.length > 0) {
    text += '    hosts {\n'
    for (const item of state.dns.hosts) {
      text += `        ${item.key}: ${item.value}\n`
    }
    text += '    }\n'
  }

  // DNS fixed_domain_ttl
  if (state.dns.fixed_domain_ttl.length > 0) {
    text += '    fixed_domain_ttl {\n'
    for (const item of state.dns.fixed_domain_ttl) {
      text += `        ${item.key}: ${item.value}\n`
    }
    text += '    }\n'
  }

  // DNS routing
  const hasDnsRouting = state.dns.request_rules.length > 0 || state.dns.response_rules.length > 0 ||
    state.dns.request_fallback !== 'asis' || state.dns.response_fallback !== 'accept'
  if (hasDnsRouting) {
    text += '    routing {\n'
    text += '        request {\n'
    for (const rule of state.dns.request_rules) {
      text += `            ${rule.rule} -> ${rule.upstream}\n`
    }
    if (state.dns.request_fallback) {
      text += `            fallback: ${state.dns.request_fallback}\n`
    }
    text += '        }\n'
    text += '        response {\n'
    for (const rule of state.dns.response_rules) {
      text += `            ${rule.rule} -> ${rule.action}\n`
    }
    if (state.dns.response_fallback) {
      text += `            fallback: ${state.dns.response_fallback}\n`
    }
    text += '        }\n'
    text += '    }\n'
  }

  text += '}\n\n'

  // Routing section
  text += sectionStart('routing')
  for (const rule of state.routing.rules) {
    const funcs = rule.functions.map(f => {
      const prefix = f.not ? '!' : ''
      const params = f.params.map(p => {
        if (!p.key) return needsQuote(p.value) ? `'${p.value}'` : p.value
        const v = needsQuote(p.value) ? `'${p.value}'` : p.value
        return `${p.key}:${v}`
      }).join(', ')
      return `${prefix}${f.name}(${params})`
    }).join(' && ')
    text += `    ${funcs} -> ${rule.outbound}\n`
  }
  text += `    fallback: ${state.routing.fallback}\n`
  text += '}\n'

  return text
}

function sectionStart(name) {
  return `${name} {\n`
}

function sectionEnd() {
  return '}\n\n'
}

function formatValue(value, key) {
  if (value === null || value === undefined) return ''
  if (value === true) return 'true'
  if (value === false) return 'false'
  if (typeof value === 'number') return String(value)
  const strVal = String(value)
  if (strVal === '') return ''
  // Quote if contains : (key-value delimiter), spaces, or special chars
  if (strVal.includes('://') || strVal.includes(' ') || strVal.includes(',') ||
      (strVal.includes(':') && !isNumPort(strVal))) {
    return `'${strVal}'`
  }
  return strVal
}

function isNumPort(s) {
  return /^\d+:\d+$/.test(s)
}

function needsQuote(s) {
  if (!s || typeof s !== 'string') return false
  return s.includes('://') || s.includes(' ') || s.includes(',') || s.includes('::') ||
    (s.includes(':') && !isNumPort(s)) || s.includes('&&') || s.startsWith('!')
}
