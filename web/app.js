async function loadJSON(path) {
  const res = await fetch(path, { headers: { 'X-Tenant-ID': 'demo' } });
  if (!res.ok) throw new Error(`${path} returned ${res.status}`);
  return res.json();
}
async function refresh() {
  try {
    const health = await loadJSON('/healthz');
    document.querySelector('#status').textContent = health.data.status || 'ok';
    document.querySelector('#tasks').innerHTML = '<li>dashboard connected</li>';
    document.querySelector('#results').innerHTML = '<li>ready</li>';
  } catch (err) {
    document.querySelector('#status').textContent = err.message;
  }
}
refresh();
