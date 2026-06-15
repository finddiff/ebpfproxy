export function parseConfig(text) {
  const state = {
    global: {},
    dns: {
      upstream: [],
      hosts: [],
      fixed_domain_ttl: [],
      request_rules: [],
      response_rules: [],
      request_fallback: 'asis',
      response_fallback: 'accept',
    },
    subscription: [],
    node: [],
    group: [],
    routing: {
      rules: [],
      fallback: 'direct'
    }
  }

  const lines = text.split('\n')
  let section = null
  let subsection = null
  let sectionDepth = -1
  let currentGroup = null

  for (let i = 0; i < lines.length; i++) {
    const line = lines[i]
    const trimmed = line.trim()
    if (!trimmed || trimmed.startsWith('#') || trimmed.startsWith('//') || trimmed.startsWith('/*')) continue

    // Track brace depth changes
    const openCount = (trimmed.match(/\{/g) || []).length
    const closeCount = (trimmed.match(/\}/g) || []).length

    // Check for section start
    if (!section && openCount > 0) {
      const secMatch = trimmed.match(/^(\w+)\s*\{/)
      if (secMatch && ['global', 'subscription', 'node', 'group', 'routing', 'dns'].includes(secMatch[1])) {
        section = secMatch[1]
        sectionDepth = 1  // We're now at depth 1 (the section body)
        subsection = null
        currentGroup = null
        continue
      }
    }

    // Inside dns section - check for nested subsection
    if (section === 'dns') {
      const subMatch = trimmed.match(/^(\w+)\s*\{/)
      if (subMatch && ['upstream', 'hosts', 'fixed_domain_ttl', 'routing', 'request', 'response'].includes(subMatch[1])) {
        subsection = subMatch[1]
        continue
      }
    }

    // Inside group section - check for named group start
    if (section === 'group') {
      const grpMatch = trimmed.match(/^([\w.\-]+)\s*\{/)
      if (grpMatch && !['upstream', 'hosts', 'routing', 'request', 'response'].includes(grpMatch[1])) {
        currentGroup = {
          name: grpMatch[1],
          policy: 'random',
          filters: [],
          tcp_check_url: '',
          tcp_check_http_method: '',
          udp_check_dns: '',
          check_interval: '',
          check_tolerance: '',
        }
        state.group.push(currentGroup)
        continue
      }
    }

    // Check for section close
    if (trimmed === '}' || trimmed === '};') {
      if (section === 'dns' && ['response', 'request', 'routing'].includes(subsection)) {
        // Go back to parent section context
        subsection = subsection === 'response' || subsection === 'request' ? 'routing' : null
        continue
      }
      // Close current level
      if (subsection && subsection !== 'routing') {
        subsection = null
        continue
      }
      if (currentGroup) {
        currentGroup = null
        continue
      }
      section = null
      subsection = null
      currentGroup = null
      sectionDepth = -1
      continue
    }

    // Parse content based on current section context
    if (section === 'global') {
      parseKeyValue(trimmed, state.global)
    } else if (section === 'subscription') {
      const val = extractQuotedValue(trimmed)
      if (val) state.subscription.push(val)
    } else if (section === 'node') {
      const nl = extractNodeLink(trimmed)
      if (nl && nl.link) state.node.push(nl.tag ? `${nl.tag}:${nl.link}` : nl.link)
    } else if (section === 'group' && currentGroup) {
      parseGroupLine(trimmed, currentGroup)
    } else if (section === 'routing') {
      parseRoutingLine(trimmed, state.routing)
    } else if (section === 'dns') {
      parseDnsLine(trimmed, state.dns, subsection)
    }
  }

  return state
}

function parseKeyValue(line, target) {
  const match = line.match(/^(\w+)\s*:\s*(.+)/)
  if (!match) return
  const key = match[1]
  let value = match[2].trim()

  if (value.endsWith(',')) value = value.slice(0, -1)

  if ((value.startsWith("'") && value.endsWith("'")) ||
      (value.startsWith('"') && value.endsWith('"'))) {
    value = value.slice(1, -1)
  }

  if (value === 'true') { target[key] = true; return }
  if (value === 'false') { target[key] = false; return }

  // Don't auto-convert number strings that might be identifiers (like port numbers)
  const num = Number(value)
  if (!isNaN(num) && value !== '' && value === String(num) && !value.startsWith('0')) {
    target[key] = num
    return
  }

  target[key] = value
}

function extractNodeLink(trimmed) {
  let val = trimmed.endsWith(',') ? trimmed.slice(0, -1) : trimmed
  // Check if it's a tagged link: "tag: 'link'" or 'tag: "link"'
  const tagMatch = val.match(/^([\w-]+)\s*:\s*(.+)/)
  if (tagMatch) {
    let link = tagMatch[2].trim()
    if ((link.startsWith("'") && link.endsWith("'")) || (link.startsWith('"') && link.endsWith('"'))) {
      link = link.slice(1, -1)
    }
    return { tag: tagMatch[1], link }
  }
  // Bare quoted link: "'link'" or '"link"'
  if ((val.startsWith("'") && val.endsWith("'")) || (val.startsWith('"') && val.endsWith('"'))) {
    val = val.slice(1, -1)
  }
  return { tag: '', link: val }
}

function parseGroupLine(trimmed, group) {
  const match = trimmed.match(/^(\w+)\s*:\s*(.+)/)
  if (!match) return
  const key = match[1]
  let value = match[2].trim()

  if (key === 'filter') {
    let val = value
    if ((val.startsWith("'") && val.endsWith("'")) || (val.startsWith('"') && val.endsWith('"'))) {
      val = val.slice(1, -1)
    }
    group.filters.push(val)
  } else if (key === 'policy') {
    if (value.endsWith(',')) value = value.slice(0, -1)
    group.policy = value
  } else if (['tcp_check_url', 'tcp_check_http_method', 'udp_check_dns', 'check_interval', 'check_tolerance'].includes(key)) {
    if ((value.startsWith("'") && value.endsWith("'")) || (value.startsWith('"') && value.endsWith('"'))) {
      value = value.slice(1, -1)
    }
    group[key] = value
  }
}

function parseRoutingLine(trimmed, routing) {
  const fbMatch = trimmed.match(/^fallback\s*:\s*(.+)/)
  if (fbMatch) {
    let val = fbMatch[1].trim()
    if ((val.startsWith("'") && val.endsWith("'")) || (val.startsWith('"') && val.endsWith('"'))) {
      val = val.slice(1, -1)
    }
    routing.fallback = val
    return
  }

  const ruleMatch = trimmed.match(/^(.+?)\s*->\s*(.+)/)
  if (!ruleMatch) return

  const matchPart = ruleMatch[1].trim()
  const outbound = ruleMatch[2].trim().replace(/,/g, '')

  const functions = []
  const parts = matchPart.split('&&').map(s => s.trim())
  for (const part of parts) {
    let not = false
    let p = part
    if (p.startsWith('!')) {
      not = true
      p = p.substring(1)
    }
    const funcMatch = p.match(/^(\w+)\((.*)\)/)
    if (funcMatch) {
      const funcName = funcMatch[1]
      const paramsStr = funcMatch[2]
      const params = []
      if (paramsStr) {
        const paramParts = splitParams(paramsStr)
        for (const pp of paramParts) {
          const kvMatch = pp.match(/^(\w+)\s*:\s*(.+)/)
          if (kvMatch) {
            let kvVal = kvMatch[2]
            if ((kvVal.startsWith("'") && kvVal.endsWith("'")) || (kvVal.startsWith('"') && kvVal.endsWith('"'))) {
              kvVal = kvVal.slice(1, -1)
            }
            params.push({ key: kvMatch[1], value: kvVal })
          } else {
            let bare = pp
            if ((bare.startsWith("'") && bare.endsWith("'")) || (bare.startsWith('"') && bare.endsWith('"'))) {
              bare = bare.slice(1, -1)
            }
            params.push({ key: '', value: bare })
          }
        }
      }
      functions.push({ name: funcName, params, not })
    }
  }

  routing.rules.push({ functions, outbound })
}

// Split params accounting for nested parens and quotes
function splitParams(str) {
  const result = []
  let depth = 0
  let inQuote = false
  let current = ''
  for (const ch of str) {
    if (inQuote) {
      current += ch
      if (ch === "'" || ch === '"') inQuote = false
      continue
    }
    if (ch === "'" || ch === '"') { inQuote = true; current += ch; continue }
    if (ch === '(') { depth++; current += ch; continue }
    if (ch === ')') { depth--; current += ch; continue }
    if (ch === ',' && depth === 0) {
      result.push(current.trim())
      current = ''
      continue
    }
    current += ch
  }
  if (current.trim()) result.push(current.trim())
  return result
}

function parseDnsLine(trimmed, dns, subsection) {
  if (subsection === 'upstream') {
    const match = trimmed.match(/^(\w+)\s*:\s*(.+)/)
    if (match) {
      let val = match[2].trim()
      if (val.endsWith(',')) val = val.slice(0, -1)
      if ((val.startsWith("'") && val.endsWith("'")) || (val.startsWith('"') && val.endsWith('"'))) {
        val = val.slice(1, -1)
      }
      dns.upstream.push({ key: match[1], value: val })
    }
  } else if (subsection === 'hosts') {
    const match = trimmed.match(/^([\w.\-]+)\s*:\s*(.+)/)
    if (match) {
      let val = match[2].trim()
      if (val.endsWith(',')) val = val.slice(0, -1)
      dns.hosts.push({ key: match[1], value: val })
    }
  } else if (subsection === 'fixed_domain_ttl') {
    const match = trimmed.match(/^([\w.\-]+)\s*:\s*(.+)/)
    if (match) {
      let val = match[2].trim()
      if (val.endsWith(',')) val = val.slice(0, -1)
      dns.fixed_domain_ttl.push({ key: match[1], value: val })
    }
  } else if (subsection === 'request') {
    const fbMatch = trimmed.match(/^fallback\s*:\s*(.+)/)
    if (fbMatch) {
      let val = fbMatch[1].trim()
      if ((val.startsWith("'") && val.endsWith("'")) || (val.startsWith('"') && val.endsWith('"'))) {
        val = val.slice(1, -1)
      }
      dns.request_fallback = val
      return
    }
    const ruleMatch = trimmed.match(/^(.+?)\s*->\s*(.+)/)
    if (ruleMatch) {
      let rule = ruleMatch[1].trim()
      let upstream = ruleMatch[2].trim()
      if ((rule.startsWith("'") && rule.endsWith("'")) || (rule.startsWith('"') && rule.endsWith('"'))) {
        rule = rule.slice(1, -1)
      }
      if ((upstream.startsWith("'") && upstream.endsWith("'")) || (upstream.startsWith('"') && upstream.endsWith('"'))) {
        upstream = upstream.slice(1, -1)
      }
      dns.request_rules.push({ rule, upstream })
    }
  } else if (subsection === 'response') {
    const fbMatch = trimmed.match(/^fallback\s*:\s*(.+)/)
    if (fbMatch) {
      let val = fbMatch[1].trim()
      if ((val.startsWith("'") && val.endsWith("'")) || (val.startsWith('"') && val.endsWith('"'))) {
        val = val.slice(1, -1)
      }
      dns.response_fallback = val
      return
    }
    const ruleMatch = trimmed.match(/^(.+?)\s*->\s*(.+)/)
    if (ruleMatch) {
      let rule = ruleMatch[1].trim()
      let action = ruleMatch[2].trim()
      if ((rule.startsWith("'") && rule.endsWith("'")) || (rule.startsWith('"') && rule.endsWith('"'))) {
        rule = rule.slice(1, -1)
      }
      dns.response_rules.push({ rule, action })
    }
  } else {
    // Top-level dns fields
    parseKeyValue(trimmed, dns)
  }
}
