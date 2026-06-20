// ── State ──
const App = {
  sessionId: null,
  direction: null,
  totalCards: 0,
  cardsRemaining: 0,
  currentIndex: 0,
  cardStates: [],
  gradeCounts: { pass: 0, hard: 0, fail: 0 },
  sessionComplete: false,
  currentCard: null,
  backRevealed: false,
  rulesData: null,
  selectedLevel: null,
  categoriesCache: null,
};

// ── API Client ──
const API = {
  base: '',

  async request(method, path, body) {
    const opts = { method, headers: {} };
    if (body) {
      opts.headers['Content-Type'] = 'application/json';
      opts.body = JSON.stringify(body);
    }
    const res = await fetch(this.base + path, opts);
    const data = await res.json();
    if (!res.ok) {
      throw new Error((data.error || 'unknown error') + ' (' + (data.code || res.status) + ')');
    }
    return data;
  },

  get(path) { return this.request('GET', path); },
  post(path, body) { return this.request('POST', path, body); },
  patch(path, body) { return this.request('PATCH', path, body); },
  put(path, body) { return this.request('PUT', path, body); },
  del(path) { return this.request('DELETE', path); },

  createSession(direction, levelNumber) {
    const body = { direction };
    if (levelNumber) body.level_number = levelNumber;
    return this.post('/api/sessions', body);
  },

  maxLevelNumber() {
    return this.get('/api/levels/max');
  },

  revealCard(sessionId) {
    return this.post('/api/sessions/' + sessionId + '/reveal');
  },

  gradeCard(sessionId, grade) {
    return this.post('/api/sessions/' + sessionId + '/grade', { grade });
  },

  listPhases() {
    return this.get('/api/phases');
  },

  createPhase(number, name) {
    return this.post('/api/phases', { number, name });
  },

  updatePhase(number, name) {
    return this.patch('/api/phases/' + number, { name });
  },

  deletePhase(number) {
    return this.del('/api/phases/' + number);
  },

  deleteLevel(number) {
    return this.del('/api/levels/' + number);
  },

  levelsInPhase(phaseNumber) {
    return this.get('/api/phases/' + phaseNumber + '/levels');
  },

  getLevel(levelNumber) {
    return this.get('/api/levels/' + levelNumber);
  },

  createLevel(data) {
    return this.post('/api/levels', data);
  },

  updateLevel(number, grammarMD) {
    return this.patch('/api/levels/' + number, {
      grammar_markdown: grammarMD,
    });
  },

  setVocabulary(number, entries) {
    return this.put('/api/levels/' + number + '/vocabulary', { vocabulary: entries });
  },

  categories() {
    return this.get('/api/levels/categories');
  },
};

// ── Session Persistence ──
const STORAGE_KEY = 'sencha-session';

function saveSessionState() {
  try {
    const state = {
      sessionId: App.sessionId,
      direction: App.direction,
      totalCards: App.totalCards,
      cardsRemaining: App.cardsRemaining,
      currentIndex: App.currentIndex,
      cardStates: App.cardStates,
      gradeCounts: App.gradeCounts,
      sessionComplete: App.sessionComplete,
      currentCard: App.currentCard,
      backRevealed: App.backRevealed,
      selectedLevel: App.selectedLevel,
    };
    sessionStorage.setItem(STORAGE_KEY, JSON.stringify(state));
  } catch (_) { /* quota exceeded, ignore */ }
}

function restoreSessionState() {
  try {
    const saved = sessionStorage.getItem(STORAGE_KEY);
    if (!saved) return;
    const state = JSON.parse(saved);
    Object.assign(App, state);
  } catch (_) { /* corrupt data, ignore */ }
}

function clearSession() {
  App.sessionId = null;
  App.direction = null;
  App.totalCards = 0;
  App.cardsRemaining = 0;
  App.currentIndex = 0;
  App.cardStates = [];
  App.gradeCounts = { pass: 0, hard: 0, fail: 0 };
  App.sessionComplete = false;
  App.currentCard = null;
  App.backRevealed = false;
  sessionStorage.removeItem(STORAGE_KEY);
}

async function verifyAndRender() {
  if (!App.sessionId) return;
  try {
    await API.get('/api/sessions/' + App.sessionId);
  } catch (err) {
    clearSession();
    location.hash = '#home?expired=1';
    return;
  }
  router();
}

// ── Error Banner ──
let errorTimeout = null;

function showError(msg) {
  const el = document.getElementById('error-banner');
  if (!el) return;
  el.textContent = msg;
  el.classList.remove('hidden');
  if (errorTimeout) clearTimeout(errorTimeout);
  errorTimeout = setTimeout(() => el.classList.add('hidden'), 5000);
}

// ── Router ──
function router() {
  const raw = location.hash || '#home';
  const hash = raw.split('?')[0];
  const params = new URLSearchParams(raw.split('?')[1] || '');
  const app = document.getElementById('app');
  if (!app) return;

  switch (hash) {
    case '#home':
      renderHome(app, params.get('expired') === '1');
      break;
    case '#level-select':
      renderLevelSelect(app);
      break;
    case '#level-picker':
      renderLevelPicker(app);
      break;
    case '#setup':
      renderSetup(app);
      break;
    case '#session':
      renderSession(app);
      break;
    case '#summary':
      renderSummary(app);
      break;
    case '#rules':
      renderRules(app);
      break;
    case '#how-it-works':
      renderHowItWorks(app);
      break;
    default:
      renderHome(app);
  }
}

window.addEventListener('hashchange', () => {
  if (location.hash.startsWith('#home')) {
    router();
  }
});
window.addEventListener('load', () => {
  restoreSessionState();
  verifyAndRender();
});

// ── Modal ──
function showModal(html) {
  closeModal();
  const overlay = document.createElement('div');
  overlay.className = 'modal-overlay';
  overlay.innerHTML = html;
  overlay.addEventListener('click', (e) => {
    if (e.target === overlay) closeModal();
  });
  document.body.appendChild(overlay);
  overlay.querySelector('.modal-close')?.addEventListener('click', closeModal);
}

function closeModal() {
  const overlay = document.querySelector('.modal-overlay');
  if (overlay) overlay.remove();
}

document.addEventListener('keydown', (e) => {
  if (e.key === 'Escape') closeModal();
});

// ── Progress Graph ──
function buildProgressGraph(cardStates, currentIndex) {
  const total = cardStates.length;
  if (total === 0) return '';
  const width = 600;
  const height = 50;
  const r = 9;
  const gap = total > 1 ? (width - 2 * r) / (total - 1) : 0;
  const cy = height / 2;
  let circles = '';
  let lines = '';

  for (let i = 0; i < total; i++) {
    const cx = total > 1 ? r + i * gap : width / 2;
    const color = cardStates[i] || '#4b5563';
    const isCurrent = i === currentIndex;
    const stroke = isCurrent ? '#fff' : 'none';
    const strokeWidth = isCurrent ? 2 : 0;
    circles += `<circle cx="${cx}" cy="${cy}" r="${r}" fill="${color}" stroke="${stroke}" stroke-width="${strokeWidth}"/>`;

    if (i < total - 1) {
      const cx2 = total > 1 ? r + (i + 1) * gap : width / 2;
      lines += `<line x1="${cx}" y1="${cy}" x2="${cx2}" y2="${cy}" stroke="#374151" stroke-width="2"/>`;
    }
  }

  return `<svg width="100%" viewBox="0 0 ${width} ${height}" class="session-progress">${lines}${circles}</svg>`;
}

// ── View: Home ──
function renderHome(app, expired) {
  const banner = expired
    ? '<div class="persistent-banner">Session expired. The server was restarted — please start a new session.</div>'
    : '';
  app.innerHTML = `
    ${banner}
    <div class="home-title">Sencha</div>
    <div class="home-buttons">
      <button class="btn btn-large btn-block" onclick="location.hash='#level-select'">[1] Start</button>
      <button class="btn btn-large btn-block" onclick="location.hash='#rules'">[2] Journey</button>
      <button class="btn btn-large btn-block" onclick="location.hash='#how-it-works'">[3] How it Works</button>
    </div>`;
}

// ── View: Level Select ──
let latestLevel = null;

async function renderLevelSelect(app) {
  app.innerHTML = '<div class="loading">Loading...</div>';
  try {
    const data = await API.maxLevelNumber();
    latestLevel = data.max;
  } catch (_) {
    latestLevel = null;
  }
  app.innerHTML = `
    <div class="home-title">Select Level</div>
    <div class="home-buttons">
      <button class="btn btn-large btn-block" onclick="startWithLatest()">[1] Latest Level${latestLevel ? ' (' + latestLevel + ')' : ''}</button>
      <button class="btn btn-large btn-block" onclick="location.hash='#level-picker'">[2] Choose a previous level</button>
      <button class="btn btn-large btn-block" onclick="location.hash='#home'" style="margin-top:24px;font-size:14px;">← Back</button>
    </div>`;
}

function startWithLatest() {
  App.selectedLevel = latestLevel || 1;
  location.hash = '#setup';
}

// ── View: Level Picker ──
async function renderLevelPicker(app) {
  app.innerHTML = '<div class="loading">Loading levels...</div>';
  try {
    const phasesData = await API.listPhases();
    const phases = phasesData.phases || [];
    const levelsByPhase = {};
    for (const phase of phases) {
      const levelsData = await API.levelsInPhase(phase.number);
      levelsByPhase[phase.number] = levelsData.levels || [];
    }

    let treeHtml = '';
    for (const phase of phases) {
      const levels = levelsByPhase[phase.number] || [];
      let levelDots = '';
      for (const level of levels) {
        levelDots += `<div class="level-node picker-level" data-level="${level.number}">${level.number}</div>`;
      }
      treeHtml += `
        <div class="phase-node active">
          <div class="phase-dot"></div>
          <span class="phase-name">${escapeHtml(phase.name)}</span>
        </div>
        <div class="phase-levels">
          <div class="level-nodes">${levelDots || '<span style="color:#6b7280;font-size:13px;">No levels</span>'}</div>
        </div>`;
    }

    app.innerHTML = `
      <div class="rules-layout">
        <div class="rules-sidebar">
          <button class="btn btn-sm" onclick="location.hash='#level-select'" style="margin-bottom:12px;">← Back</button>
          ${treeHtml || '<div style="color:#6b7280;padding:12px;">No levels yet</div>'}
        </div>
        <div class="rules-content">
          <div style="color:#6b7280;text-align:center;padding:60px 0;">Select a level to start a session</div>
        </div>
      </div>`;

    app.querySelectorAll('.picker-level').forEach(el => {
      el.addEventListener('click', () => {
        const levelNum = parseInt(el.dataset.level);
        showLevelDetailForPicker(levelNum);
      });
    });
  } catch (err) {
    showError('Failed to load levels: ' + err.message);
    app.innerHTML = '';
  }
}

async function showLevelDetailForPicker(levelNumber) {
  try {
    const data = await API.getLevel(levelNumber);
    const level = data.level;
    const vocab = data.vocabulary || [];

    let grammarHtml = level.grammar_md ? marked.parse(level.grammar_md) : '<em>No grammar</em>';

    let vocabHtml = '';
    if (vocab.length === 0) {
      vocabHtml = '<em style="color:#6b7280;">No vocabulary for this level</em>';
    } else {
      vocabHtml = '<table class="vocab-table"><thead><tr><th>Korean</th><th>English</th><th>Category</th></tr></thead><tbody>';
      for (const v of vocab) {
        vocabHtml += `<tr><td class="korean">${escapeHtml(v.korean)}</td><td class="english">${escapeHtml(v.english)}</td><td class="category">${escapeHtml(v.category) || ''}</td></tr>`;
      }
      vocabHtml += '</tbody></table>';
    }

    showModal(`
      <div class="modal">
        <button class="modal-close">&times;</button>
        <div class="modal-title">Level ${level.number}</div>
        <div class="modal-body">
          <div class="modal-body-col">
            <h3>Grammar</h3>
            <div class="grammar-content">${grammarHtml}</div>
          </div>
          <div class="modal-body-col">
            <h3>Vocabulary</h3>
            ${vocabHtml}
          </div>
        </div>
        <div class="form-actions">
          <button class="btn btn-sm" onclick="closeModal()">Cancel</button>
          <button class="btn btn-sm btn-green" onclick="startFromPicker(${level.number})">Start</button>
        </div>
      </div>`);
  } catch (err) {
    showError('Failed to load level: ' + err.message);
  }
}

function startFromPicker(levelNumber) {
  closeModal();
  App.selectedLevel = levelNumber;
  location.hash = '#setup';
}

// ── View: How it Works ──
function renderHowItWorks(app) {
  app.innerHTML = `
    <div style="width:100%;text-align:left;margin-bottom:12px;"><button class="btn btn-sm" onclick="location.hash='#home'">← Home</button></div>
    <div style="max-width:640px;margin:0 auto;text-align:left;">

      <h1 style="color:#fff;font-size:28px;margin:0 0 24px;">How Sencha Works</h1>

      <h2 style="color:#4ade80;font-size:18px;margin:0 0 8px;">Phases &amp; Levels</h2>
      <p style="color:#d1d5db;font-size:15px;line-height:1.7;margin:0 0 24px;">
        The curriculum is split into <strong>phases</strong> (broad topic groups).
        Each phase contains numbered <strong>levels</strong>. Every level introduces
        new grammar rules and vocabulary.
      </p>

      <h2 style="color:#4ade80;font-size:18px;margin:0 0 8px;">Sentence Generation</h2>
      <p style="color:#d1d5db;font-size:15px;line-height:1.7;margin:0 0 8px;">
        When you start a session for a level, the LLM generates practice sentences using:
      </p>
      <ul style="color:#d1d5db;font-size:15px;line-height:1.7;margin:0 0 24px;padding-left:20px;">
        <li><strong>Grammar</strong> — only the current level's rules</li>
        <li><strong>Vocabulary</strong> — 50 randomly chosen words from across all levels (or all words if fewer than 50 exist)</li>
      </ul>

      <h2 style="color:#4ade80;font-size:18px;margin:0 0 8px;">Grading</h2>
      <p style="color:#d1d5db;font-size:15px;line-height:1.7;margin:0 0 24px;">
        After revealing a card, rate yourself:<br>
        <span style="color:#4ade80;">Pass</span> — you knew it &nbsp;|&nbsp;
        <span style="color:#fbbf24;">Hard</span> — you struggled &nbsp;|&nbsp;
        <span style="color:#f87171;">Fail</span> — you didn't know it
      </p>

      <h2 style="color:#4ade80;font-size:18px;margin:0 0 8px;">Keyboard Shortcuts</h2>
      <p style="color:#d1d5db;font-size:15px;line-height:1.7;margin:0 0 24px;">
        <code style="background:#1f2937;padding:1px 6px;border-radius:3px;">1</code>/<code style="background:#1f2937;padding:1px 6px;border-radius:3px;">2</code>/<code style="background:#1f2937;padding:1px 6px;border-radius:3px;">3</code> grade &nbsp;|&nbsp;
        <code style="background:#1f2937;padding:1px 6px;border-radius:3px;">ENTER</code>/<code style="background:#1f2937;padding:1px 6px;border-radius:3px;">SPACE</code> reveal &nbsp;|&nbsp;
        <code style="background:#1f2937;padding:1px 6px;border-radius:3px;">ESC</code> close &nbsp;|&nbsp;
        <code style="background:#1f2937;padding:1px 6px;border-radius:3px;">S</code> new session &nbsp;|&nbsp;
        <code style="background:#1f2937;padding:1px 6px;border-radius:3px;">Q</code> quit
      </p>

    </div>`;
}

// ── View: Session Setup ──
let selectedDirection = 'korean-to-english';

function renderSetup(app) {
  selectedDirection = 'korean-to-english';
  app.innerHTML = `
    <div style="width:100%;text-align:left;margin-bottom:12px;"><button class="btn btn-sm" onclick="location.hash='#home'">← Home</button></div>
    <div class="setup-title">Select Direction</div>
    <div class="setup-options">
      <div class="setup-option selected" data-dir="korean-to-english">
        <span class="shortcut">[1]</span> Korean → English
      </div>
      <div class="setup-option" data-dir="english-to-korean">
        <span class="shortcut">[2]</span> English → Korean
      </div>
      <div class="setup-option" data-dir="mixed">
        <span class="shortcut">[3]</span> Mixed
      </div>
    </div>
    <button class="btn btn-green btn-large" onclick="startSession()">Start Session</button>`;

  app.querySelectorAll('.setup-option').forEach(el => {
    el.addEventListener('click', () => {
      app.querySelectorAll('.setup-option').forEach(o => o.classList.remove('selected'));
      el.classList.add('selected');
      selectedDirection = el.dataset.dir;
    });
  });
}

async function startSession() {
  try {
    const data = await API.createSession(selectedDirection, App.selectedLevel);
    App.sessionId = data.session_id;
    App.direction = data.direction;
    App.totalCards = data.total_cards;
    App.cardsRemaining = data.cards_remaining;
    App.currentIndex = 0;
    App.sessionComplete = false;
    App.gradeCounts = { pass: 0, hard: 0, fail: 0 };
    App.currentCard = null;
    App.backRevealed = false;
    App.cardStates = new Array(data.total_cards).fill('#4b5563');
    saveSessionState();
    await revealCard();
    location.hash = '#session';
  } catch (err) {
    showError('Failed to create session: ' + err.message);
  }
}

// ── View: Session ──
function renderSession(app) {
  if (!App.sessionId) {
    location.hash = '#home';
    return;
  }
  const completed = App.currentIndex;
  const total = App.totalCards;
  const graph = buildProgressGraph(App.cardStates, App.currentIndex);

  let cardHtml;
  let actionsHtml;

  if (!App.currentCard) {
    cardHtml = '<div class="card-front" style="color:#6b7280;">Loading card...</div>';
    actionsHtml = '';
  } else if (!App.backRevealed) {
    cardHtml = renderCardFront(App.currentCard);
    actionsHtml = '<button class="btn btn-green btn-large" onclick="backReveal()" id="reveal-btn">Reveal</button>';
  } else {
    cardHtml = renderCardContent(App.currentCard);
    actionsHtml = renderGradeButtons();
  }

  app.innerHTML = `
    ${graph}
    <div class="session-stats">
      <span class="stat-completed">Completed: ${completed} / ${total}</span>
      <span class="stat-pass">Pass: ${App.gradeCounts.pass}</span>
      <span class="stat-hard">Hard: ${App.gradeCounts.hard}</span>
      <span class="stat-fail">Fail: ${App.gradeCounts.fail}</span>
    </div>
    <div class="session-card" id="session-card">${cardHtml}</div>
    <div class="session-actions" id="session-actions">${actionsHtml}</div>`;
}

function renderCardFront(card) {
  return `
    <div class="card-front">${escapeHtml(card.front)}</div>
    <div class="card-divider"></div>
    <div class="card-back" style="visibility:hidden">${escapeHtml(card.back)}</div>`;
}

function renderCardContent(reveal) {
  return `
    <div class="card-front">${escapeHtml(reveal.front)}</div>
    <div class="card-divider"></div>
    <div class="card-back">${escapeHtml(reveal.back)}</div>`;
}

function renderGradeButtons() {
  return `
    <button class="btn btn-green" onclick="gradeCard('pass')">[1] Pass</button>
    <button class="btn btn-gold" onclick="gradeCard('hard')">[2] Hard</button>
    <button class="btn btn-red" onclick="gradeCard('fail')">[3] Fail</button>`;
}

async function revealCard() {
  try {
    const data = await API.revealCard(App.sessionId);
    App.currentCard = data;
    App.backRevealed = false;
    App.cardStates[App.currentIndex] = '#6b7280';
    saveSessionState();
    const app = document.getElementById('app');
    if (app) renderSession(app);
  } catch (err) {
    if (err.message.includes('NOT_FOUND')) {
      clearSession();
      location.hash = '#home?expired=1';
      return;
    }
    showError('Failed to reveal card: ' + err.message);
  }
}

function backReveal() {
  App.backRevealed = true;
  saveSessionState();
  const app = document.getElementById('app');
  if (app) renderSession(app);
}

async function gradeCard(grade) {
  try {
    const data = await API.gradeCard(App.sessionId, grade);
    App.gradeCounts[grade]++;
    App.cardsRemaining = data.cards_remaining;
    App.sessionComplete = data.session_complete;
    const color = grade === 'pass' ? '#4ade80' : grade === 'hard' ? '#fbbf24' : '#f87171';
    App.cardStates[App.currentIndex] = color;

    if (data.session_complete) {
      clearSession();
      location.hash = '#summary';
    } else {
      App.currentIndex++;
      App.currentCard = null;
      App.backRevealed = false;
      saveSessionState();
      await revealCard();
    }
  } catch (err) {
    if (err.message.includes('NOT_FOUND')) {
      clearSession();
      location.hash = '#home?expired=1';
      return;
    }
    showError('Failed to grade card: ' + err.message);
  }
}

// ── View: Summary ──
function renderSummary(app) {
  const total = App.totalCards;
  const passed = App.gradeCounts.pass;
  const hard = App.gradeCounts.hard;
  const failed = App.gradeCounts.fail;
  const percent = total > 0 ? Math.round((passed / total) * 100) : 0;

  app.innerHTML = `
    <div class="summary-title">Session Complete!</div>
    <div class="summary-stats">
      <div class="summary-stat pass"><div class="count">${passed}</div><div class="label">Pass</div></div>
      <div class="summary-stat hard"><div class="count">${hard}</div><div class="label">Hard</div></div>
      <div class="summary-stat fail"><div class="count">${failed}</div><div class="label">Fail</div></div>
    </div>
    <div class="summary-percent">${percent}%</div>
    <div class="summary-buttons">
      <button class="btn btn-green btn-large" onclick="clearSession();location.hash='#setup'">[S] Start New</button>
      <button class="btn btn-large" onclick="clearSession();location.hash='#home'">[Q] Quit</button>
    </div>`;
}

// ── View: Rules ──
async function renderRules(app) {
  app.innerHTML = '<div class="loading">Loading rules...</div>';
  try {
    const phasesData = await API.listPhases();
    const phases = phasesData.phases || [];
    const levelsByPhase = {};

    for (const phase of phases) {
      const levelsData = await API.levelsInPhase(phase.number);
      levelsByPhase[phase.number] = levelsData.levels || [];
    }

    App.rulesData = { phases, levelsByPhase };
    renderRulesContent(app);
  } catch (err) {
    showError('Failed to load rules: ' + err.message);
    app.innerHTML = '';
  }
}

function renderRulesContent(app) {
  const { phases, levelsByPhase } = App.rulesData;
  let sidebarHtml = '<button class="btn btn-sm" onclick="location.hash=\'#home\'" style="margin-bottom:12px;">← Home</button>';

  for (const phase of phases) {
    const levels = levelsByPhase[phase.number] || [];
    let levelDots = '';
    for (const level of levels) {
      levelDots += `<div class="level-node" data-level="${level.number}">${level.number}</div>`;
    }
    sidebarHtml += `
      <div class="phase-node active">
        <div class="phase-dot"></div>
        <span class="phase-name">${escapeHtml(phase.name)}</span>
        <span class="phase-edit" data-phase="${phase.number}" data-name="${escapeHtml(phase.name)}" onclick="showEditPhaseModal(this.dataset.phase, this.dataset.name)" style="margin-left:auto;cursor:pointer;font-size:14px;color:#6b7280;">&#9998;</span>
        <span data-phase="${phase.number}" data-name="${escapeHtml(phase.name)}" style="cursor:pointer;font-size:14px;color:#f87171;margin-left:8px;" onclick="confirmDeletePhase(this.dataset.phase, this.dataset.name)">&#128465;</span>
      </div>
      <div class="phase-levels">
        <div class="level-nodes">${levelDots || '<span style="color:#6b7280;font-size:13px;">No levels</span>'}</div>
      </div>`;
  }

  app.innerHTML = `
    <div class="rules-layout">
      <div class="rules-sidebar">
        ${sidebarHtml || '<div style="color:#6b7280;padding:12px;">No phases yet</div>'}
        <div class="rules-actions">
          <button class="btn btn-sm" onclick="showAddPhaseModal()">+ Add Phase</button>
          <button class="btn btn-sm" onclick="showAddLevelModal()">+ Add Level</button>
        </div>
      </div>
      <div class="rules-content">
        <div style="color:#6b7280;text-align:center;padding:60px 0;">Select a level to view rules</div>
      </div>
    </div>`;

  app.querySelectorAll('.level-node').forEach(el => {
    el.addEventListener('click', () => {
      const levelNum = parseInt(el.dataset.level);
      showLevelDetail(levelNum);
    });
  });
}

async function showLevelDetail(levelNumber) {
  try {
    const data = await API.getLevel(levelNumber);
    const level = data.level;
    const vocab = data.level_vocabulary || [];

    let grammarHtml = level.grammar_md ? marked.parse(level.grammar_md) : '<em>No grammar</em>';

    let vocabHtml = '';
    if (vocab.length === 0) {
      vocabHtml = '<em style="color:#6b7280;">No vocabulary for this level</em>';
    } else {
      vocabHtml = '<table class="vocab-table"><thead><tr><th>Korean</th><th>English</th><th>Category</th></tr></thead><tbody>';
      for (const v of vocab) {
        vocabHtml += `<tr><td class="korean">${escapeHtml(v.korean)}</td><td class="english">${escapeHtml(v.english)}</td><td class="category">${escapeHtml(v.category) || ''}</td></tr>`;
      }
      vocabHtml += '</tbody></table>';
    }

    showModal(`
      <div class="modal">
        <button class="modal-close">&times;</button>
        <div class="modal-title">Level ${level.number}</div>
        <div class="modal-body">
          <div class="modal-body-col">
            <h3>Grammar</h3>
            <div class="grammar-content">${grammarHtml}</div>
          </div>
          <div class="modal-body-col">
            <h3>Vocabulary</h3>
            ${vocabHtml}
          </div>
        </div>
        <div class="form-actions">
          <button class="btn btn-sm" onclick="showEditLevelForm(${level.number})">Edit Rules</button>
          <button class="btn btn-sm" onclick="showEditVocabForm(${level.number})">Edit Vocab</button>
          <button class="btn btn-sm" style="background:#b91c1c;" onclick="confirmDeleteLevel(${level.number})">Delete Level</button>
        </div>
      </div>`);
  } catch (err) {
    showError('Failed to load level: ' + err.message);
  }
}

async function showEditLevelForm(levelNumber) {
  try {
    closeModal();
    const data = await API.getLevel(levelNumber);
    const level = data.level;

    showModal(`
      <div class="modal">
        <button class="modal-close">&times;</button>
        <div class="modal-title">Edit Level ${level.number} Rules</div>
        <div class="modal-form">
          <label>Grammar (Markdown)</label>
          <textarea id="edit-level-grammar">${escapeHtml(level.grammar_md || '')}</textarea>
          <div class="form-actions">
            <button class="btn btn-sm" onclick="closeModal()">Cancel</button>
            <button class="btn btn-sm btn-green" onclick="submitEditLevel(${levelNumber})">Save</button>
          </div>
        </div>
      </div>`);
  } catch (err) {
    showError('Failed to load level: ' + err.message);
  }
}

async function submitEditLevel(levelNumber) {
  const grammar = document.getElementById('edit-level-grammar').value.trim();
  try {
    await API.updateLevel(levelNumber, grammar);
    closeModal();
    showLevelDetail(levelNumber);
  } catch (err) {
    showError('Failed to update level: ' + err.message);
  }
}

// ── Modal: Add Phase ──
function showAddPhaseModal() {
  const next = (App.rulesData?.phases?.length || 0) + 1;
  showModal(`
    <div class="modal">
      <button class="modal-close">&times;</button>
      <div class="modal-title">Add Phase</div>
      <div class="modal-form">
        <label>Phase Number</label>
        <input type="number" id="add-phase-number" min="1" value="${next}">
        <label>Phase Name</label>
        <input type="text" id="add-phase-name" placeholder="e.g. Phase ${next}">
        <div class="form-actions">
          <button class="btn btn-sm" onclick="closeModal()">Cancel</button>
          <button class="btn btn-sm btn-green" onclick="submitAddPhase()">Create</button>
        </div>
      </div>
    </div>`);
}

async function submitAddPhase() {
  const number = parseInt(document.getElementById('add-phase-number').value);
  const name = document.getElementById('add-phase-name').value.trim();
  if (!name) { showError('Phase name is required'); return; }
  try {
    await API.createPhase(number, name);
    closeModal();
    renderRules(document.getElementById('app'));
  } catch (err) {
    showError('Failed to create phase: ' + err.message);
  }
}

// ── Modal: Edit Phase ──
function showEditPhaseModal(number, name) {
  showModal(`
    <div class="modal">
      <button class="modal-close">&times;</button>
      <div class="modal-title">Edit Phase ${number}</div>
      <div class="modal-form">
        <label>Phase Number</label>
        <input type="number" id="edit-phase-number" value="${number}" disabled>
        <label>Phase Name</label>
        <input type="text" id="edit-phase-name" value="${name}">
        <div class="form-actions">
          <button class="btn btn-sm" onclick="closeModal()">Cancel</button>
          <button class="btn btn-sm btn-green" onclick="submitEditPhase()">Save</button>
        </div>
      </div>
    </div>`);
}

async function submitEditPhase() {
  const number = parseInt(document.getElementById('edit-phase-number').value);
  const name = document.getElementById('edit-phase-name').value.trim();
  if (!name) { showError('Phase name is required'); return; }
  try {
    await API.updatePhase(number, name);
    closeModal();
    renderRules(document.getElementById('app'));
  } catch (err) {
    showError('Failed to update phase: ' + err.message);
  }
}

// ── Modal: Edit Vocabulary ──
async function showEditVocabForm(levelNumber) {
  try {
    closeModal();
    const data = await API.getLevel(levelNumber);
    const vocab = data.level_vocabulary || [];

    await ensureCategories();

      let rowsHtml = '';
    if (vocab.length === 0) {
      rowsHtml = `
        <div class="vocab-row">
          <input type="text" class="vocab-korean" placeholder="Korean">
          <input type="text" class="vocab-english" placeholder="English">
          <input class="vocab-category" list="cd" placeholder="noun">
          <button class="btn-remove" onclick="this.parentElement.remove()">&times;</button>
        </div>`;
    } else {
      for (const v of vocab) {
        rowsHtml += `
          <div class="vocab-row">
            <input type="text" class="vocab-korean" value="${escapeHtml(v.korean)}">
            <input type="text" class="vocab-english" value="${escapeHtml(v.english)}">
            <input class="vocab-category" list="cd" value="${escapeHtml(v.category)}" placeholder="noun">
            <button class="btn-remove" onclick="this.parentElement.remove()">&times;</button>
          </div>`;
      }
    }

    showModal(`
      <div class="modal">
        <button class="modal-close">&times;</button>
        <div class="modal-title">Edit Level ${levelNumber} Vocabulary</div>
        <datalist id="cd">${categoryOptionsHtml()}</datalist>
        <div class="modal-form">
          <div id="edit-vocab-rows">${rowsHtml}</div>
          <button class="btn btn-sm" onclick="addEditVocabRow()" style="margin-top:8px;">+ Add another word</button>
          <div class="form-actions">
            <button class="btn btn-sm" onclick="showLevelDetail(${levelNumber})">Cancel</button>
            <button class="btn btn-sm btn-green" onclick="submitEditVocab(${levelNumber})">Save</button>
          </div>
        </div>
      </div>`);
  } catch (err) {
    showError('Failed to load vocabulary: ' + err.message);
  }
}

function addEditVocabRow() {
  const container = document.getElementById('edit-vocab-rows');
  const row = document.createElement('div');
  row.className = 'vocab-row';
  row.innerHTML = `
    <input type="text" class="vocab-korean" placeholder="Korean">
    <input type="text" class="vocab-english" placeholder="English">
    <input class="vocab-category" list="cd" placeholder="noun">
    <button class="btn-remove" onclick="this.parentElement.remove()">&times;</button>`;
  container.appendChild(row);
}

async function submitEditVocab(levelNumber) {
  const rows = document.querySelectorAll('#edit-vocab-rows .vocab-row');
  const vocabulary = [];
  for (const row of rows) {
    const korean = row.querySelector('.vocab-korean').value.trim();
    const english = row.querySelector('.vocab-english').value.trim();
    const category = row.querySelector('.vocab-category').value;
    if (korean && english) {
      vocabulary.push({ korean, english, category });
    }
  }
  try {
    await API.setVocabulary(levelNumber, vocabulary);
    showLevelDetail(levelNumber);
  } catch (err) {
    showError('Failed to save vocabulary: ' + err.message);
  }
}

// ── Confirm Delete ──
function confirmDeletePhase(number, name) {
  const { levelsByPhase } = App.rulesData;
  const levelCount = (levelsByPhase[parseInt(number)] || []).length;
  const warning = levelCount > 0
    ? `<p style="color:#f87171;margin-top:8px;">This will permanently delete <strong>${levelCount} level${levelCount > 1 ? 's' : ''}</strong> in this phase.</p>`
    : '';
  showModal(`
    <div class="modal" style="max-width:420px;">
      <button class="modal-close">&times;</button>
      <div class="modal-title" style="color:#f87171;">Delete Phase?</div>
      <div class="modal-body">
        <p>Are you sure you want to delete <strong>${escapeHtml(name)}</strong>?</p>
        ${warning}
        <p style="margin-top:8px;font-size:13px;color:#6b7280;">This action cannot be undone.</p>
      </div>
      <div class="form-actions">
        <button class="btn btn-sm" onclick="closeModal()">Cancel</button>
        <button class="btn btn-sm" style="background:#b91c1c;" onclick="submitDeletePhase(${number})">Delete Phase</button>
      </div>
    </div>`);
}

async function submitDeletePhase(number) {
  closeModal();
  try {
    await API.deletePhase(number);
    renderRules(document.getElementById('app'));
  } catch (err) {
    showError('Failed to delete phase: ' + err.message);
  }
}

function confirmDeleteLevel(number) {
  const { phases, levelsByPhase } = App.rulesData;
  let renumberCount = 0;
  for (const phase of phases) {
    const levels = levelsByPhase[phase.number] || [];
    renumberCount = levels.filter(l => l.number > number).length;
    if (renumberCount > 0) break;
  }
  const warning = renumberCount > 0
    ? `<p style="color:#fbbf24;margin-top:8px;"><strong>${renumberCount} level${renumberCount > 1 ? 's' : ''}</strong> after this one will be renumbered.</p>`
    : '';
  showModal(`
    <div class="modal" style="max-width:420px;">
      <button class="modal-close">&times;</button>
      <div class="modal-title" style="color:#f87171;">Delete Level?</div>
      <div class="modal-body">
        <p>Are you sure you want to delete <strong>Level ${number}</strong>?</p>
        ${warning}
        <p style="margin-top:8px;font-size:13px;color:#6b7280;">This action cannot be undone.</p>
      </div>
      <div class="form-actions">
        <button class="btn btn-sm" onclick="closeModal()">Cancel</button>
        <button class="btn btn-sm" style="background:#b91c1c;" onclick="submitDeleteLevel(${number})">Delete Level</button>
      </div>
    </div>`);
}

async function submitDeleteLevel(number) {
  closeModal();
  try {
    await API.deleteLevel(number);
    renderRules(document.getElementById('app'));
  } catch (err) {
    showError('Failed to delete level: ' + err.message);
  }
}

// ── Modal: Add Level ──
async function showAddLevelModal() {
  let phaseOptions = '';
  const { phases } = App.rulesData;
  for (const phase of phases) {
    phaseOptions += `<option value="${phase.number}">${escapeHtml(phase.name)}</option>`;
  }

  await ensureCategories();

  showModal(`
    <div class="modal">
      <button class="modal-close">&times;</button>
      <div class="modal-title">Add Level</div>
      <datalist id="cd">${categoryOptionsHtml()}</datalist>
      <div class="modal-form" id="add-level-form">
        <label>Phase</label>
        <select id="add-level-phase">${phaseOptions}</select>
        <label>Grammar Rules (Markdown)</label>
        <textarea id="add-level-grammar" placeholder="Enter grammar rules in markdown..."></textarea>
        <label>Vocabulary</label>
        <div id="vocab-rows">
          <div class="vocab-row">
            <input type="text" class="vocab-korean" placeholder="Korean">
            <input type="text" class="vocab-english" placeholder="English">
            <input class="vocab-category" list="cd" placeholder="noun">
            <button class="btn-remove" onclick="this.parentElement.remove()">&times;</button>
          </div>
        </div>
        <button class="btn btn-sm" onclick="addVocabRow()" style="margin-top:8px;">+ Add another word</button>
        <div class="form-actions">
          <button class="btn btn-sm" onclick="closeModal()">Cancel</button>
          <button class="btn btn-sm btn-green" onclick="submitAddLevel()">Create</button>
        </div>
      </div>
    </div>`);
}

function addVocabRow() {
  const container = document.getElementById('vocab-rows');
  const row = document.createElement('div');
  row.className = 'vocab-row';
  row.innerHTML = `
    <input type="text" class="vocab-korean" placeholder="Korean">
    <input type="text" class="vocab-english" placeholder="English">
    <input class="vocab-category" list="cd" placeholder="noun">
    <button class="btn-remove" onclick="this.parentElement.remove()">&times;</button>`;
  container.appendChild(row);
}

async function submitAddLevel() {
  const phaseNumber = parseInt(document.getElementById('add-level-phase').value);
  const grammar = document.getElementById('add-level-grammar').value.trim();

  if (!grammar) { showError('Grammar rules are required'); return; }

  const vocabRows = document.querySelectorAll('#vocab-rows .vocab-row');
  const vocabulary = [];
  for (const row of vocabRows) {
    const korean = row.querySelector('.vocab-korean').value.trim();
    const english = row.querySelector('.vocab-english').value.trim();
    const category = row.querySelector('.vocab-category').value;
    if (korean && english) {
      vocabulary.push({ korean, english, category });
    }
  }

  try {
    await API.createLevel({
      phase_number: phaseNumber,
      grammar_markdown: grammar,
      vocabulary,
    });
    closeModal();
    renderRules(document.getElementById('app'));
  } catch (err) {
    showError('Failed to create level: ' + err.message);
  }
}

// ── Keyboard Handling ──
document.addEventListener('keydown', (e) => {
  if (e.key === 'Escape') { closeModal(); return; }

  const hash = location.hash || '#home';

  if (hash === '#home') {
    if (e.key === '1') location.hash = '#level-select';
    if (e.key === '2') location.hash = '#rules';
    if (e.key === '3') { location.hash = '#how-it-works'; e.preventDefault(); }
  }

  if (hash === '#level-select') {
    if (e.key === '1') { startWithLatest(); e.preventDefault(); }
    if (e.key === '2') { location.hash = '#level-picker'; e.preventDefault(); }
  }

  if (hash === '#setup') {
    if (e.key === '1') {
      document.querySelector('[data-dir="korean-to-english"]')?.click();
      e.preventDefault();
    }
    if (e.key === '2') {
      document.querySelector('[data-dir="english-to-korean"]')?.click();
      e.preventDefault();
    }
    if (e.key === '3') {
      document.querySelector('[data-dir="mixed"]')?.click();
      e.preventDefault();
    }
    if (e.key === 'Enter') {
      document.querySelector('.btn-green')?.click();
      e.preventDefault();
    }
  }

  if (hash === '#session') {
    const revealBtn = document.getElementById('reveal-btn');
    if (revealBtn && (e.key === 'Enter' || e.key === ' ')) {
      revealBtn.click();
      e.preventDefault();
    }

    if (!revealBtn) {
      if (e.key === '1') {
        document.querySelector('.btn-green')?.click();
        e.preventDefault();
      }
      if (e.key === '2') {
        document.querySelector('.btn-gold')?.click();
        e.preventDefault();
      }
      if (e.key === '3') {
        document.querySelector('.btn-red')?.click();
        e.preventDefault();
      }
    }
  }

  if (hash === '#summary') {
    if (e.key === 's' || e.key === 'S') location.hash = '#setup';
    if (e.key === 'q' || e.key === 'Q') location.hash = '#home';
    if (e.key === 'Enter') e.preventDefault();
  }
});

// ── Helper ──
function escapeHtml(str) {
  if (!str) return '';
  const div = document.createElement('div');
  div.textContent = str;
  return div.innerHTML;
}

let categoriesPromise = null;

function categoryOptionsHtml(selected) {
  if (!App.categoriesCache) return '';
  return App.categoriesCache.map(c =>
    `<option value="${escapeHtml(c)}"${c === selected ? ' selected' : ''}>${escapeHtml(c)}</option>`
  ).join('');
}

async function ensureCategories() {
  if (App.categoriesCache) return;
  if (!categoriesPromise) {
    categoriesPromise = API.categories().then(d => {
      App.categoriesCache = d.categories || [];
    }).catch(() => {
      App.categoriesCache = [];
    });
  }
  await categoriesPromise;
}
