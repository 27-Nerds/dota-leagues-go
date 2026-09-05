<script>
  import PageHeader from "./PageHeader.svelte";
  import Leagues from "./Leagues.svelte";
  import LeagueDetail from "./LeagueDetail.svelte";
  import MatchPage from "./MatchPage.svelte";
  import TeamPage from "./TeamPage.svelte";
  import Teams from "./Teams.svelte";
  import Updates from "./Updates.svelte";
  import DpcPage from "./DpcPage.svelte";
  import StatePanel from "./StatePanel.svelte";
  import { parsePath } from "./lib/routes.js";
  const route = parsePath(window.location.pathname);
  const hasServerMetadata = !!document.querySelector('meta[name="description"]');
  const descriptions = {
    updates: 'Follow recorded Dota 2 tournament changes, team records, roster moves, and newly added players.',
    home: 'Browse Dota 2 leagues, tournament schedules, live games, and match results.',
    teams: 'Explore Dota 2 teams, player rosters, statistics, and tournament history.',
    team: 'View a Dota 2 team roster, player statistics, and tournament history.',
    league: 'Explore this Dota 2 league’s schedule, live games, and tournament results.',
    match: 'View a Dota 2 match scoreboard, teams, heroes, and player statistics.',
    dpc: 'Explore official Dota Pro Circuit standings and league results.'
  };
</script>

<svelte:head>{#if !hasServerMetadata && descriptions[route.name]}<meta name="description" content={descriptions[route.name]} />{/if}{#if route.name === "notFound"}<title>Page not found | Dota 2 Leagues</title><meta name="robots" content="noindex" />{/if}</svelte:head>

<a class="skip-link" href="#main">Skip to content</a>
<header>
  <div class="header-inner">
    <a class="brand" href="/" aria-label="Dota 2 Leagues home"><img src="/logo.png" alt="Dota 2 Leagues" /></a>
    <nav aria-label="Main navigation">
      <a href="/" aria-current={['home','league','match'].includes(route.name) ? 'page' : undefined}>Leagues</a>
      <a href="/team" aria-current={['teams','team'].includes(route.name) ? 'page' : undefined}>Teams</a>
      <a href="/activity" aria-current={route.name === 'updates' ? 'page' : undefined}>Updates</a>
      <a href="/dpc" aria-current={route.name === 'dpc' ? 'page' : undefined}>DPC Standings</a>
    </nav>
  </div>
</header>
<main id="main" tabindex="-1">
  {#if route.name === "home"}<Leagues />
  {:else if route.name === "league"}<LeagueDetail id={route.id} />
  {:else if route.name === "match"}<MatchPage leagueId={route.leagueId} matchId={route.matchId} />
  {:else if route.name === "team"}<TeamPage id={route.id} />
  {:else if route.name === "teams"}<Teams />
  {:else if route.name === "updates"}<Updates />
  {:else if route.name === "dpc"}<DpcPage />
  {:else}

    <PageHeader title="Page not found" description="The page you’re looking for is unavailable." />
    <StatePanel kind="notFound" title="We couldn’t find that page" message="The link may be incomplete or the page may have moved. Browse leagues to find a tournament." actionLabel="Browse leagues" href="/" />
  {/if}
</main>
<footer><span>Dota 2 Leagues</span><p>Tournaments, teams, and the games between them.</p><a href="/">Browse leagues</a><a href="/team">Explore teams</a><a href="/activity">Latest updates</a></footer>

<style>
  header { background: var(--surface); border-bottom: 1px solid var(--line); }
  .header-inner { max-width: 1280px; margin: auto; padding: 16px 28px; display: flex; align-items: center; justify-content: space-between; gap: 24px; }
  .brand { display: flex; flex: 0 0 129px; min-width: 0; }
  .brand img { width: 100%; max-width: 129px; height: auto; filter: brightness(.45); }
  nav { display: flex; gap: 6px; flex-wrap: wrap; }
  nav a { text-decoration: none; color: var(--muted); font-size: 13px; padding: 10px 15px; white-space: nowrap; }
  nav a:hover { color: var(--text); }
  nav a[aria-current] { color: var(--text); font-weight: 600; }
  main { max-width: 1280px; min-width: 0; min-height: calc(100dvh - 190px); margin: auto; padding: 32px 28px 48px; font-size: 14px; line-height: 1.6; }
  main:focus { outline: none; }
  footer { border-top: 1px solid var(--line); padding: 24px 28px; display: flex; gap: 20px; align-items: center; justify-content: center; flex-wrap: wrap; font-size: 12px; color: var(--muted); background: var(--surface); }
  footer span { color: var(--text); font-weight: 600; } footer p { margin: 0; }
  .skip-link { position: fixed; left: 12px; top: -60px; padding: 12px; background: var(--gold); z-index: 20; }
  .skip-link:focus { top: 12px; }
  @media(max-width: 600px) {
    .header-inner { padding: 16px; flex-direction: column; gap: 14px; align-items: flex-start; }
    .brand { flex: auto; width: 129px; }
    nav { width: 100%; gap: 4px; } nav a { flex: 1; text-align: center; padding: 10px 8px; }
    main { padding: 26px 16px 44px; }
    footer { justify-content: flex-start; gap: 12px 20px; } footer p { flex-basis: 100%; }
  }
</style>
