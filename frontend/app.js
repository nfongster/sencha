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
  lastReveal: null,
  rulesData: null,
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

  createSession(direction) {
    return this.post('/api/sessions', { direction });
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

  levelsInPhase(phaseNumber) {
    return this.get('/api/phases/' + phaseNumber + '/levels');
  },

  getLevel(levelNumber) {
    return this.get('/api/levels/' + levelNumber);
  },

  createLevel(data) {
    return this.post('/api/levels', data);
  },
};

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
  const hash = location.hash || '#home';
  const app = document.getElementById('app');
  if (!app) return;

  switch (hash) {
    case '#home':
      renderHome(app);
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
    default:
      renderHome(app);
  }
}

window.addEventListener('hashchange', router);
window.addEventListener('load', router);

// ── Modal ──
function showModal(html) {
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
function renderHome(app) {
  app.innerHTML = `
    <div class="home-title">Sencha</div>
    <div class="home-buttons">
      <button class="btn btn-large btn-block" onclick="location.hash='#setup'">[1] Start</button>
      <button class="btn btn-large btn-block" onclick="location.hash='#rules'">[2] Rules</button>
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
    const data = await API.createSession(selectedDirection);
    App.sessionId = data.session_id;
    App.direction = data.direction;
    App.totalCards = data.total_cards;
    App.cardsRemaining = data.cards_remaining;
    App.currentIndex = 0;
    App.sessionComplete = false;
    App.gradeCounts = { pass: 0, hard: 0, fail: 0 };
    App.lastReveal = null;
    App.cardStates = new Array(data.total_cards).fill('#4b5563');
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

  const cardHtml = App.lastReveal
    ? renderCardContent(App.lastReveal)
    : '<div class="card-front" style="color:#6b7280;">Press Reveal to see the card</div>';

  const actionsHtml = App.lastReveal
    ? renderGradeButtons()
    : '<button class="btn btn-green btn-large" onclick="revealCard()" id="reveal-btn">Reveal</button>';

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
    App.lastReveal = data;
    App.cardStates[App.currentIndex] = '#6b7280';
    const app = document.getElementById('app');
    if (app) renderSession(app);
  } catch (err) {
    showError('Failed to reveal card: ' + err.message);
  }
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
      location.hash = '#summary';
    } else {
      App.currentIndex++;
      App.lastReveal = null;
      const app = document.getElementById('app');
      if (app) renderSession(app);
    }
  } catch (err) {
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
      <button class="btn btn-green btn-large" onclick="location.hash='#setup'">[S] Start New</button>
      <button class="btn btn-large" onclick="location.hash='#home'">[Q] Quit</button>
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
    const vocab = data.vocabulary || [];

    let grammarHtml = level.grammar_md ? marked.parse(level.grammar_md) : '<em>No grammar</em>';
    if (level.exceptions_md) {
      grammarHtml += '<h3>Exceptions</h3>' + marked.parse(level.exceptions_md);
    }

    let vocabHtml = '';
    if (vocab.length === 0) {
      vocabHtml = '<em style="color:#6b7280;">No vocabulary for this level</em>';
    } else {
      vocabHtml = '<ul class="vocab-list">';
      for (const v of vocab) {
        vocabHtml += `<li><span class="korean">${escapeHtml(v.korean)}</span><span class="english">${escapeHtml(v.english)}</span></li>`;
      }
      vocabHtml += '</ul>';
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
      </div>`);
  } catch (err) {
    showError('Failed to load level: ' + err.message);
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
    location.hash = '#rules';
  } catch (err) {
    showError('Failed to create phase: ' + err.message);
  }
}

// ── Modal: Add Level ──
async function showAddLevelModal() {
  let phaseOptions = '';
  const { phases } = App.rulesData;
  for (const phase of phases) {
    phaseOptions += `<option value="${phase.number}">${escapeHtml(phase.name)}</option>`;
  }

  showModal(`
    <div class="modal">
      <button class="modal-close">&times;</button>
      <div class="modal-title">Add Level</div>
      <div class="modal-form" id="add-level-form">
        <label>Phase</label>
        <select id="add-level-phase">${phaseOptions}</select>
        <label>Grammar Rules (Markdown)</label>
        <textarea id="add-level-grammar" placeholder="Enter grammar rules in markdown..."></textarea>
        <label>Exceptions (optional, Markdown)</label>
        <textarea id="add-level-exceptions" placeholder="Enter exceptions in markdown..."></textarea>
        <label>Vocabulary</label>
        <div id="vocab-rows">
          <div class="vocab-row">
            <input type="text" class="vocab-korean" placeholder="Korean">
            <input type="text" class="vocab-english" placeholder="English">
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
    <button class="btn-remove" onclick="this.parentElement.remove()">&times;</button>`;
  container.appendChild(row);
}

async function submitAddLevel() {
  const phaseNumber = parseInt(document.getElementById('add-level-phase').value);
  const grammar = document.getElementById('add-level-grammar').value.trim();
  const exceptions = document.getElementById('add-level-exceptions').value.trim();

  if (!grammar) { showError('Grammar rules are required'); return; }

  const vocabRows = document.querySelectorAll('#vocab-rows .vocab-row');
  const vocabulary = [];
  for (const row of vocabRows) {
    const korean = row.querySelector('.vocab-korean').value.trim();
    const english = row.querySelector('.vocab-english').value.trim();
    if (korean && english) {
      vocabulary.push({ korean, english });
    }
  }

  try {
    await API.createLevel({
      phase_number: phaseNumber,
      grammar_markdown: grammar,
      exceptions_markdown: exceptions || '',
      vocabulary,
    });
    closeModal();
    location.hash = '#rules';
  } catch (err) {
    showError('Failed to create level: ' + err.message);
  }
}

// ── Keyboard Handling ──
document.addEventListener('keydown', (e) => {
  if (e.key === 'Escape') { closeModal(); return; }

  const hash = location.hash || '#home';

  if (hash === '#home') {
    if (e.key === '1') location.hash = '#setup';
    if (e.key === '2') location.hash = '#rules';
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
