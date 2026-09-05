const BASE_URL = process.env.baseUrl;

export async function api(path, params) {
  let url = `${BASE_URL}${path}`;
  if (params) {
    const qs = Object.keys(params)
      .filter(k => params[k] !== undefined && params[k] !== null)
      .map(k => `${encodeURIComponent(k)}=${encodeURIComponent(params[k])}`)
      .join("&");
    if (qs) url += `?${qs}`;
  }

  let res;
  try {
    res = await fetch(url);
  } catch (e) {
    throw new Error("We couldn’t connect. Check your connection and try again.");
  }

  if (res.status === 429) {
    const error = new Error("Too many requests. Wait a moment, then try again.");
    error.status = 429;
    throw error;
  }

  let data;
  const text = await res.text();
  try {
    data = text ? JSON.parse(text) : null;
  } catch (e) {
    data = text;
  }

  if (!res.ok) {
    const msg = typeof data === "string" && data ? data
      : (data && data.message) || `server error ${res.status}`;
    const error = new Error(msg);
    error.status = res.status;
    throw error;
  }

  return data;
}

export async function apiResults(path, params) {
  const body = await api(path, params);
  if (!body || !Array.isArray(body.results)) {
    throw new Error("unexpected response shape");
  }
  return { meta: { ...body.meta, total: Number.isFinite(Number(body.meta?.total)) ? Number(body.meta.total) : body.results.length }, results: body.results };
}
