// Consume once, before filters rewrite the URL. Missing data (e.g. Vite dev)
// falls back to the normal API request.
export function takeDirectoryData(path, doc = document, location = window.location) {
  const node = doc.getElementById('directory-data');
  if (!node) return null;
  node.remove();
  try {
    const data = JSON.parse(node.textContent);
    if (data.path !== path || data.search !== location.search.replace(/^\?/, '') ||
        !Array.isArray(data.results) || !Number.isFinite(data.meta?.total)) return null;
    return data;
  } catch {
    return null;
  }
}
