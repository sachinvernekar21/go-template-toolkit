package tt

import (
	"strings"
	"testing"
)

func TestHTMLTableWithLoopVars(t *testing.T) {
	tmpl := `<table>
<thead><tr><th>#</th><th>Name</th><th>Email</th></tr></thead>
<tbody>
[% FOREACH user IN users %]
<tr class="[% IF loop.index mod 2 == 0 %]even[% ELSE %]odd[% END %][% IF loop.first %] first[% END %][% IF loop.last %] last[% END %]">
  <td>[% loop.count %]</td>
  <td>[% user.name | html %]</td>
  <td>[% user.email | html %]</td>
</tr>
[% END %]
</tbody>
</table>`

	vars := map[string]interface{}{
		"users": []interface{}{
			map[string]interface{}{"name": "Alice <Admin>", "email": "alice@example.com"},
			map[string]interface{}{"name": "Bob & Co", "email": "bob@example.com"},
			map[string]interface{}{"name": `Carol "CJ" Jones`, "email": "carol@example.com"},
		},
	}

	result := evalTemplate(t, tmpl, vars)

	if !strings.Contains(result, `Alice &lt;Admin&gt;`) {
		t.Error("expected HTML-escaped angle brackets in Alice's name")
	}
	if !strings.Contains(result, `Bob &amp; Co`) {
		t.Error("expected HTML-escaped ampersand in Bob's name")
	}
	if !strings.Contains(result, `Carol &quot;CJ&quot; Jones`) {
		t.Error("expected HTML-escaped quotes in Carol's name")
	}
	if !strings.Contains(result, `class="even first"`) {
		t.Error("expected first row to have 'even first' class")
	}
	if !strings.Contains(result, `class="odd"`) {
		t.Error("expected second row to have 'odd' class")
	}
	if !strings.Contains(result, `class="even last"`) {
		t.Error("expected third row to have 'even last' class")
	}
	if strings.Count(result, "<td>") != 9 {
		t.Errorf("expected 9 <td> cells (3 rows x 3 cols), got %d", strings.Count(result, "<td>"))
	}
}

func TestHTMLNavMenuActiveState(t *testing.T) {
	tmpl := `<nav>
<ul>
[% FOREACH page IN pages %]
  <li[% IF page.slug == active %] class="active"[% END %]><a href="/[% page.slug %]">[% page.title | html %]</a></li>
[% END %]
</ul>
</nav>`

	vars := map[string]interface{}{
		"active": "about",
		"pages": []interface{}{
			map[string]interface{}{"slug": "home", "title": "Home"},
			map[string]interface{}{"slug": "about", "title": "About Us"},
			map[string]interface{}{"slug": "contact", "title": "Contact & Info"},
		},
	}

	result := evalTemplate(t, tmpl, vars)

	if !strings.Contains(result, `<li><a href="/home">Home</a></li>`) {
		t.Error("expected Home link without active class")
	}
	if !strings.Contains(result, `<li class="active"><a href="/about">About Us</a></li>`) {
		t.Error("expected About link with active class")
	}
	if !strings.Contains(result, `Contact &amp; Info`) {
		t.Error("expected HTML-escaped ampersand in Contact title")
	}
	if strings.Count(result, `class="active"`) != 1 {
		t.Errorf("expected exactly 1 active item, got %d", strings.Count(result, `class="active"`))
	}
}

func TestHTMLProductCardGrid(t *testing.T) {
	tmpl := `<div class="grid">
[% FOREACH product IN products %]
<div class="card">
  <span class="num">#[% loop.count %]</span>
  <h3>[% product.name | html | truncate(20) %]</h3>
  <p class="price">$[% product.price %]</p>
  <span class="badge [% product.in_stock ? 'in-stock' : 'sold-out' %]">
    [% product.in_stock ? 'In Stock' : 'Sold Out' %]
  </span>
  <p class="tax">Tax: $[% product.price * 0.1 %]</p>
</div>
[% END %]
</div>`

	vars := map[string]interface{}{
		"products": []interface{}{
			map[string]interface{}{"name": "Wireless Bluetooth Speaker Pro Max", "price": 49.99, "in_stock": true},
			map[string]interface{}{"name": "USB-C Hub", "price": 29.99, "in_stock": false},
		},
	}

	result := evalTemplate(t, tmpl, vars)

	if !strings.Contains(result, "#1") || !strings.Contains(result, "#2") {
		t.Error("expected product numbering with loop.count")
	}
	if !strings.Contains(result, "Wireless Bluetooth S...") {
		t.Error("expected truncated product name for long title")
	}
	if !strings.Contains(result, "USB-C Hub") {
		t.Error("expected short product name unchanged")
	}
	if !strings.Contains(result, `class="badge in-stock"`) {
		t.Error("expected in-stock badge for first product")
	}
	if !strings.Contains(result, `class="badge sold-out"`) {
		t.Error("expected sold-out badge for second product")
	}
	if !strings.Contains(result, "In Stock") {
		t.Error("expected 'In Stock' text")
	}
	if !strings.Contains(result, "Sold Out") {
		t.Error("expected 'Sold Out' text")
	}
}

func TestHTMLFormDynamicFields(t *testing.T) {
	tmpl := `<form action="/submit" method="post">
[% FOREACH field IN fields %]
  <div class="field">
    <label for="[% field.name %]">[% field.label | html %]</label>
    [% SWITCH field.type %]
    [% CASE 'text' %]
    <input type="text" id="[% field.name %]" name="[% field.name %]" value="[% DEFAULT field.value = '' %][% field.value | html %]">
    [% CASE 'textarea' %]
    <textarea id="[% field.name %]" name="[% field.name %]">[% field.value | html %]</textarea>
    [% CASE 'select' %]
    <select id="[% field.name %]" name="[% field.name %]">
      [% FOREACH opt IN field.options %]
      <option value="[% opt.value | html %]"[% IF opt.value == field.value %] selected[% END %]>[% opt.label | html %]</option>
      [% END %]
    </select>
    [% CASE %]
    <input type="text" id="[% field.name %]" name="[% field.name %]">
    [% END %]
  </div>
[% END %]
<button type="submit">Submit</button>
</form>`

	vars := map[string]interface{}{
		"fields": []interface{}{
			map[string]interface{}{
				"name": "username", "label": "User Name", "type": "text", "value": "john\"doe",
			},
			map[string]interface{}{
				"name": "bio", "label": "Biography", "type": "textarea", "value": "Hello <world>",
			},
			map[string]interface{}{
				"name": "role", "label": "Role", "type": "select", "value": "editor",
				"options": []interface{}{
					map[string]interface{}{"value": "viewer", "label": "Viewer"},
					map[string]interface{}{"value": "editor", "label": "Editor"},
					map[string]interface{}{"value": "admin", "label": "Admin"},
				},
			},
		},
	}

	result := evalTemplate(t, tmpl, vars)

	if !strings.Contains(result, `value="john&quot;doe"`) {
		t.Error("expected HTML-escaped quotes in text input value")
	}
	if !strings.Contains(result, `Hello &lt;world&gt;`) {
		t.Error("expected HTML-escaped content in textarea")
	}
	if !strings.Contains(result, `<option value="editor" selected>Editor</option>`) {
		t.Error("expected selected attribute on editor option")
	}
	if strings.Count(result, " selected") != 1 {
		t.Errorf("expected exactly 1 selected option, got %d", strings.Count(result, " selected"))
	}
	if !strings.Contains(result, `<button type="submit">Submit</button>`) {
		t.Error("expected submit button")
	}
}

func TestHTMLEmailWithMacros(t *testing.T) {
	tmpl := `[% MACRO heading(text) BLOCK %]<h2 style="color: #333; font-family: Arial;">[% text | html %]</h2>[% END %]
[% MACRO button(url, label) BLOCK %]<a href="[% url %]" style="background: #007bff; color: #fff; padding: 10px 20px; text-decoration: none; border-radius: 4px;">[% label | html %]</a>[% END %]
<div style="max-width: 600px; margin: 0 auto; font-family: Arial;">
  [% heading('Welcome, ' _ user.name _ '!') %]
  <p>Thank you for signing up on [% site_name | html %].</p>
  <p>Your account has been created with the email: <strong>[% user.email | html %]</strong></p>
  [% IF user.is_premium %]
  <p style="color: green;">You have premium access!</p>
  [% END %]
  <p>[% button('/dashboard', 'Go to Dashboard') %]</p>
  <hr>
  <p style="font-size: 12px; color: #999;">This email was sent to [% user.email %]. If you didn't sign up, please ignore.</p>
</div>`

	vars := map[string]interface{}{
		"site_name": "My App & Co",
		"user": map[string]interface{}{
			"name":       "Alice",
			"email":      "alice@example.com",
			"is_premium": true,
		},
	}

	result := evalTemplate(t, tmpl, vars)

	if !strings.Contains(result, `Welcome, Alice!`) {
		t.Error("expected macro heading with concatenated name")
	}
	if !strings.Contains(result, `<h2 style="color: #333; font-family: Arial;">`) {
		t.Error("expected styled heading from macro")
	}
	if !strings.Contains(result, `My App &amp; Co`) {
		t.Error("expected HTML-escaped site name")
	}
	if !strings.Contains(result, `You have premium access!`) {
		t.Error("expected premium message for premium user")
	}
	if !strings.Contains(result, `href="/dashboard"`) {
		t.Error("expected dashboard link from button macro")
	}
	if !strings.Contains(result, `Go to Dashboard`) {
		t.Error("expected button label from macro")
	}
}

func TestHTMLBlogPostXSS(t *testing.T) {
	tmpl := `<article>
  <h1>[% post.title | html %]</h1>
  <div class="meta">By [% post.author | html %] on [% post.date %]</div>
  <div class="content">
    [% post.body | html | html_para %]
  </div>
  [% IF post.tags.size %]
  <div class="tags">
    [% FOREACH tag IN post.tags %]
    <span class="tag">[% tag | html %]</span>
    [% END %]
  </div>
  [% END %]
</article>`

	vars := map[string]interface{}{
		"post": map[string]interface{}{
			"title":  `<script>alert("xss")</script>`,
			"author": `Evil"User<br>`,
			"date":   "2024-01-15",
			"body":   "First paragraph with <b>bold</b> attempt.\n\nSecond paragraph with <script>evil()</script> code.\n\nThird paragraph.",
			"tags":   []interface{}{"go", `<img onerror=alert(1)>`, "templates"},
		},
	}

	result := evalTemplate(t, tmpl, vars)

	if strings.Contains(result, "<script>") {
		t.Error("XSS: unescaped <script> tag found in output")
	}
	if strings.Contains(result, "<img ") {
		t.Error("XSS: unescaped <img> tag found in output")
	}
	if !strings.Contains(result, `&lt;script&gt;alert(&quot;xss&quot;)&lt;/script&gt;`) {
		t.Error("expected fully escaped script tag in title")
	}
	if !strings.Contains(result, `Evil&quot;User&lt;br&gt;`) {
		t.Error("expected escaped author name")
	}
	if !strings.Contains(result, `&lt;img onerror=alert(1)&gt;`) {
		t.Error("expected escaped img tag in tags section")
	}
	if strings.Count(result, "<p>") != 3 {
		t.Errorf("expected 3 paragraphs from html_para, got %d", strings.Count(result, "<p>"))
	}
	if strings.Count(result, `class="tag"`) != 3 {
		t.Errorf("expected 3 tag spans, got %d", strings.Count(result, `class="tag"`))
	}
}

func TestHTMLLayoutWithWrapper(t *testing.T) {
	engine := New()
	engine.SetLoader(NewStringLoader(map[string]string{
		"layout": `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>[% page_title | html %]</title>
  [% BLOCK head_extra %][% END %]
</head>
<body>
  <header>
    [% INCLUDE nav %]
  </header>
  <main>
    [% content %]
  </main>
  <footer><p>&copy; 2024 [% site_name | html %]</p></footer>
</body>
</html>`,
		"nav": `<nav><a href="/">Home</a> | <a href="/about">About</a></nav>`,
	}))

	tmpl := `[% WRAPPER layout page_title='Dashboard' %]
<h1>Welcome, [% user_name | html %]</h1>
<p>You have [% message_count %] new messages.</p>
[% IF message_count > 0 %]
<ul>
[% FOREACH msg IN messages %]
  <li>[% msg | html %]</li>
[% END %]
</ul>
[% END %]
[% END %]`

	vars := map[string]interface{}{
		"site_name":     "Acme Corp",
		"user_name":     "Jane <Doe>",
		"message_count": 2,
		"messages":      []interface{}{"Hello & welcome!", "Meeting at 3pm"},
	}

	result, err := engine.ProcessString(tmpl, vars)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(result, "<!DOCTYPE html>") {
		t.Error("expected DOCTYPE from wrapper layout")
	}
	if !strings.Contains(result, "<title>Dashboard</title>") {
		t.Error("expected page title in <title> tag")
	}
	if !strings.Contains(result, `Jane &lt;Doe&gt;`) {
		t.Error("expected HTML-escaped user name")
	}
	if !strings.Contains(result, `Hello &amp; welcome!`) {
		t.Error("expected HTML-escaped message")
	}
	if !strings.Contains(result, `<nav><a href="/">Home</a>`) {
		t.Error("expected nav from INCLUDE")
	}
	if !strings.Contains(result, `&copy; 2024 Acme Corp`) {
		t.Error("expected footer with site name")
	}
	if strings.Count(result, "<li>") != 2 {
		t.Errorf("expected 2 message list items, got %d", strings.Count(result, "<li>"))
	}
}

func TestHTMLNestedConditionalSections(t *testing.T) {
	tmpl := `<div class="dashboard">
  <h1>Dashboard</h1>
  [% IF user.role == 'admin' %]
  <section class="admin-panel">
    <h2>Admin Panel</h2>
    [% IF user.permissions.manage_users %]
    <div class="widget">
      <h3>User Management</h3>
      <p>[% user_count %] users registered</p>
    </div>
    [% END %]
    [% IF user.permissions.view_logs %]
    <div class="widget">
      <h3>System Logs</h3>
      <p>Last entry: [% last_log | html | truncate(50) %]</p>
    </div>
    [% END %]
  </section>
  [% ELSIF user.role == 'editor' %]
  <section class="editor-panel">
    <h2>Editor Panel</h2>
    <p>You can edit content.</p>
  </section>
  [% ELSE %]
  <section class="viewer-panel">
    <h2>Viewer Panel</h2>
    <p>Read-only access.</p>
  </section>
  [% END %]
  [% UNLESS user.email_verified %]
  <div class="alert alert-warning">
    <p>Please verify your email: [% user.email | html %]</p>
  </div>
  [% END %]
</div>`

	adminVars := map[string]interface{}{
		"user": map[string]interface{}{
			"role": "admin",
			"permissions": map[string]interface{}{
				"manage_users": true,
				"view_logs":    true,
			},
			"email":          "admin@example.com",
			"email_verified": false,
		},
		"user_count": 150,
		"last_log":   "2024-01-15 12:30:00 - System update completed successfully and all checks passed",
	}

	result := evalTemplate(t, tmpl, adminVars)

	if !strings.Contains(result, `class="admin-panel"`) {
		t.Error("expected admin panel for admin role")
	}
	if strings.Contains(result, `class="editor-panel"`) {
		t.Error("should not show editor panel for admin")
	}
	if !strings.Contains(result, "User Management") {
		t.Error("expected user management widget for admin with manage_users permission")
	}
	if !strings.Contains(result, "150 users registered") {
		t.Error("expected user count")
	}
	if !strings.Contains(result, "System Logs") {
		t.Error("expected system logs widget for admin with view_logs permission")
	}
	if !strings.Contains(result, "...") {
		t.Error("expected truncated log entry")
	}
	if !strings.Contains(result, `class="alert alert-warning"`) {
		t.Error("expected email verification warning for unverified user")
	}

	editorVars := map[string]interface{}{
		"user": map[string]interface{}{
			"role":           "editor",
			"email":          "editor@example.com",
			"email_verified": true,
		},
	}

	result = evalTemplate(t, tmpl, editorVars)

	if !strings.Contains(result, `class="editor-panel"`) {
		t.Error("expected editor panel for editor role")
	}
	if strings.Contains(result, `class="admin-panel"`) {
		t.Error("should not show admin panel for editor")
	}
	if strings.Contains(result, `class="alert alert-warning"`) {
		t.Error("should not show email warning for verified user")
	}
}

func TestHTMLDefinitionListFromHash(t *testing.T) {
	tmpl := `<dl class="metadata">
[% FOREACH key IN info.sort %]
  <dt>[% key | upper %]</dt>
  <dd>[% info.item(key) | html %]</dd>
[% END %]
</dl>
<p>Total fields: [% info.size %]</p>`

	vars := map[string]interface{}{
		"info": map[string]interface{}{
			"version":  "2.0",
			"author":   "Jane & John",
			"language": "Go",
		},
	}

	result := evalTemplate(t, tmpl, vars)

	if !strings.Contains(result, "<dt>AUTHOR</dt>") {
		t.Error("expected uppercased 'AUTHOR' key")
	}
	if !strings.Contains(result, `Jane &amp; John`) {
		t.Error("expected HTML-escaped author value")
	}
	if !strings.Contains(result, "<dt>LANGUAGE</dt>") {
		t.Error("expected 'LANGUAGE' key")
	}
	if !strings.Contains(result, "<dt>VERSION</dt>") {
		t.Error("expected 'VERSION' key")
	}

	authorIdx := strings.Index(result, "AUTHOR")
	langIdx := strings.Index(result, "LANGUAGE")
	verIdx := strings.Index(result, "VERSION")
	if !(authorIdx < langIdx && langIdx < verIdx) {
		t.Error("expected keys sorted alphabetically: author < language < version")
	}

	if !strings.Contains(result, "Total fields: 3") {
		t.Error("expected hash size of 3")
	}
}

func TestHTMLErrorPageTryCatch(t *testing.T) {
	tmpl := `<!DOCTYPE html>
<html>
<body>
[% TRY %]
  <div class="content">
    <h1>User Profile</h1>
    [% THROW 'db' 'Connection refused to database server' %]
    <p>This should not appear</p>
  </div>
[% CATCH db %]
  <div class="error error-db">
    <h1>Database Error</h1>
    <p class="error-type">Error type: [% error.type %]</p>
    <p class="error-detail">[% error.info | html %]</p>
    <p>Please try again later or contact support.</p>
  </div>
[% CATCH %]
  <div class="error error-generic">
    <h1>Unexpected Error</h1>
    <p>Something went wrong.</p>
  </div>
[% FINAL %]
  <footer class="error-footer">
    <p><a href="/">Return to Home</a> | <a href="/support">Contact Support</a></p>
  </footer>
[% END %]
</body>
</html>`

	result := evalTemplate(t, tmpl, nil)

	if strings.Contains(result, "This should not appear") {
		t.Error("content after THROW should not render")
	}
	if !strings.Contains(result, `class="error error-db"`) {
		t.Error("expected db-specific error section")
	}
	if strings.Contains(result, `class="error error-generic"`) {
		t.Error("should not show generic error when db error is caught")
	}
	if !strings.Contains(result, "Error type: db") {
		t.Error("expected error.type to be 'db'")
	}
	if !strings.Contains(result, "Connection refused to database server") {
		t.Error("expected error.info message")
	}
	if !strings.Contains(result, `class="error-footer"`) {
		t.Error("expected FINAL footer to always render")
	}
	if !strings.Contains(result, `<a href="/">Return to Home</a>`) {
		t.Error("expected home link in FINAL footer")
	}
	if !strings.Contains(result, "<!DOCTYPE html>") {
		t.Error("expected DOCTYPE at top of page")
	}
}

func TestMultiLevelDeepDotAccess(t *testing.T) {
	tmpl := `<div class="org">
<h1>[% company.name | html %]</h1>
[% FOREACH dept IN company.departments %]
<div class="dept">
  <h2>[% dept.name %]</h2>
  [% FOREACH team IN dept.teams %]
  <div class="team">
    <h3>[% team.name %] (Lead: [% team.lead.name | html %])</h3>
    <ul>
    [% FOREACH member IN team.members %]
      <li>[% member.name | html %] &lt;[% member.email | html %]&gt; - [% member.role %]</li>
    [% END %]
    </ul>
  </div>
  [% END %]
</div>
[% END %]
</div>`

	vars := map[string]interface{}{
		"company": map[string]interface{}{
			"name": "Acme & Sons",
			"departments": []interface{}{
				map[string]interface{}{
					"name": "Engineering",
					"teams": []interface{}{
						map[string]interface{}{
							"name": "Backend",
							"lead": map[string]interface{}{"name": "Alice <CTO>"},
							"members": []interface{}{
								map[string]interface{}{"name": "Bob", "email": "bob@acme.com", "role": "Senior"},
								map[string]interface{}{"name": "Carol & Dave", "email": "cd@acme.com", "role": "Junior"},
							},
						},
						map[string]interface{}{
							"name": "Frontend",
							"lead": map[string]interface{}{"name": "Eve"},
							"members": []interface{}{
								map[string]interface{}{"name": "Frank", "email": "frank@acme.com", "role": "Mid"},
							},
						},
					},
				},
				map[string]interface{}{
					"name": "Sales",
					"teams": []interface{}{
						map[string]interface{}{
							"name": "Enterprise",
							"lead": map[string]interface{}{"name": "Grace"},
							"members": []interface{}{
								map[string]interface{}{"name": "Hank", "email": "hank@acme.com", "role": "Lead"},
							},
						},
					},
				},
			},
		},
	}

	result := evalTemplate(t, tmpl, vars)

	if !strings.Contains(result, "Acme &amp; Sons") {
		t.Error("expected HTML-escaped company name at level 1")
	}
	if !strings.Contains(result, "<h2>Engineering</h2>") {
		t.Error("expected department name at level 2")
	}
	if !strings.Contains(result, "<h2>Sales</h2>") {
		t.Error("expected Sales department at level 2")
	}
	if !strings.Contains(result, "Alice &lt;CTO&gt;") {
		t.Error("expected HTML-escaped team lead at level 3 (team.lead.name)")
	}
	if !strings.Contains(result, "Carol &amp; Dave") {
		t.Error("expected HTML-escaped member name at level 4")
	}
	if !strings.Contains(result, "bob@acme.com") {
		t.Error("expected member email at level 4")
	}
	if strings.Count(result, `class="dept"`) != 2 {
		t.Errorf("expected 2 departments, got %d", strings.Count(result, `class="dept"`))
	}
	if strings.Count(result, `class="team"`) != 3 {
		t.Errorf("expected 3 teams total, got %d", strings.Count(result, `class="team"`))
	}
	if strings.Count(result, "<li>") != 4 {
		t.Errorf("expected 4 members total, got %d", strings.Count(result, "<li>"))
	}
}

func TestMultiLevelTripleNestedLoop(t *testing.T) {
	tmpl := `[% FOREACH cat IN categories %]
<section class="category[% IF loop.first %] first-cat[% END %]">
  <h1>[% loop.count %]. [% cat.name %]</h1>
  [% SET cat_count = loop.count %]
  [% FOREACH sub IN cat.subs %]
  <div class="sub[% IF loop.last %] last-sub[% END %]">
    <h2>[% cat_count %].[% loop.count %] [% sub.name %]</h2>
    [% SET sub_count = loop.count %]
    <ul>
    [% FOREACH item IN sub.items %]
      <li class="[% IF loop.first %]first-item[% END %][% IF loop.last %]last-item[% END %]">[% cat_count %].[% sub_count %].[% loop.count %] [% item | html %]</li>
    [% END %]
    </ul>
  </div>
  [% END %]
</section>
[% END %]`

	vars := map[string]interface{}{
		"categories": []interface{}{
			map[string]interface{}{
				"name": "Fruit",
				"subs": []interface{}{
					map[string]interface{}{
						"name":  "Citrus",
						"items": []interface{}{"Orange", "Lemon"},
					},
					map[string]interface{}{
						"name":  "Berries",
						"items": []interface{}{"Strawberry", "Blueberry", "Raspberry"},
					},
				},
			},
			map[string]interface{}{
				"name": "Vegetables",
				"subs": []interface{}{
					map[string]interface{}{
						"name":  "Leafy",
						"items": []interface{}{"Spinach"},
					},
				},
			},
		},
	}

	result := evalTemplate(t, tmpl, vars)

	if !strings.Contains(result, `class="category first-cat"`) {
		t.Error("expected first-cat class on first category (outer loop.first)")
	}
	if !strings.Contains(result, "1. Fruit") {
		t.Error("expected outer loop.count=1 for Fruit")
	}
	if !strings.Contains(result, "2. Vegetables") {
		t.Error("expected outer loop.count=2 for Vegetables")
	}
	if !strings.Contains(result, `class="sub last-sub"`) {
		t.Error("expected last-sub class on last sub-category (middle loop.last)")
	}
	if !strings.Contains(result, "1.1 Citrus") {
		t.Error("expected cat_count.sub_count = 1.1 for Citrus")
	}
	if !strings.Contains(result, "1.2 Berries") {
		t.Error("expected cat_count.sub_count = 1.2 for Berries")
	}
	if !strings.Contains(result, `class="first-item"`) {
		t.Error("expected first-item class on first item (inner loop.first)")
	}
	if !strings.Contains(result, `class="last-item"`) {
		t.Error("expected last-item class on last item (inner loop.last)")
	}
	if !strings.Contains(result, "1.1.1 Orange") {
		t.Error("expected triple numbering 1.1.1 Orange")
	}
	if !strings.Contains(result, "1.2.3 Raspberry") {
		t.Error("expected triple numbering 1.2.3 Raspberry")
	}
	if !strings.Contains(result, "2.1.1 Spinach") {
		t.Error("expected triple numbering 2.1.1 Spinach")
	}
}

func TestMultiLevelTemplateComposition(t *testing.T) {
	engine := New()
	engine.SetLoader(NewStringLoader(map[string]string{
		"page_layout": `<!DOCTYPE html>
<html>
<head><title>[% page_title | html %]</title></head>
<body>
  <aside>[% INCLUDE sidebar %]</aside>
  <main>[% content %]</main>
  <footer>[% INCLUDE footer_block %]</footer>
</body>
</html>`,
		"sidebar": `<nav>
[% FOREACH link IN nav_links %]
  <a href="[% link.url %]">[% link.label | html %]</a>
[% END %]
</nav>`,
		"footer_block": `<p>&copy; [% year %] [% site_name | html %]</p>`,
	}))

	tmpl := `[% MACRO badge(text, color) BLOCK %]<span class="badge" style="color:[% color %]">[% text | html %]</span>[% END %]
[% WRAPPER page_layout page_title='User Profile' %]
<h1>[% user.name | html %]</h1>
<p>Role: [% badge(user.role, 'blue') %]</p>
<p>Status: [% badge(user.status, user.status == 'active' ? 'green' : 'red') %]</p>
[% IF user.projects.size %]
<h2>Projects</h2>
<ul>
[% FOREACH proj IN user.projects %]
  <li>[% proj | html %]</li>
[% END %]
</ul>
[% END %]
[% END %]`

	vars := map[string]interface{}{
		"site_name": "DevHub",
		"year":      2024,
		"user": map[string]interface{}{
			"name":     "Jane <Dev>",
			"role":     "Admin",
			"status":   "active",
			"projects": []interface{}{"Alpha", "Beta & Gamma"},
		},
		"nav_links": []interface{}{
			map[string]interface{}{"url": "/home", "label": "Home"},
			map[string]interface{}{"url": "/profile", "label": "Profile"},
		},
	}

	result, err := engine.ProcessString(tmpl, vars)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(result, "<!DOCTYPE html>") {
		t.Error("expected DOCTYPE from WRAPPER layout")
	}
	if !strings.Contains(result, "<title>User Profile</title>") {
		t.Error("expected page title from WRAPPER param")
	}
	if !strings.Contains(result, `<a href="/home">Home</a>`) {
		t.Error("expected nav link from INCLUDE sidebar")
	}
	if !strings.Contains(result, `Jane &lt;Dev&gt;`) {
		t.Error("expected HTML-escaped user name in wrapped content")
	}
	if !strings.Contains(result, `class="badge"`) {
		t.Error("expected badge from MACRO")
	}
	if !strings.Contains(result, `style="color:green"`) {
		t.Error("expected green color from ternary in MACRO arg")
	}
	if !strings.Contains(result, "Beta &amp; Gamma") {
		t.Error("expected HTML-escaped project in FOREACH inside WRAPPER")
	}
	if !strings.Contains(result, "&copy; 2024 DevHub") {
		t.Error("expected footer from INCLUDE inside WRAPPER layout")
	}
}

func TestMultiLevelConditionalNesting(t *testing.T) {
	tmpl := `<div class="access-control">
[% IF user.org %]
  <h1>[% user.org.name | html %]</h1>
  [% IF user.org.department %]
    <h2>[% user.org.department.name %]</h2>
    [% IF user.org.department.permissions %]
      [% IF user.org.department.permissions.can_approve %]
      <div class="action approve">
        <button>Approve Requests</button>
        <p>Budget limit: $[% user.org.department.permissions.budget_limit %]</p>
      </div>
      [% ELSIF user.org.department.permissions.can_view %]
      <div class="action view-only">
        <p>You have view-only access.</p>
      </div>
      [% ELSE %]
      <div class="action no-access">
        <p>No permissions configured.</p>
      </div>
      [% END %]
    [% ELSE %]
    <div class="warning">No permissions object found.</div>
    [% END %]
  [% ELSE %]
  <div class="warning">No department assigned.</div>
  [% END %]
[% ELSE %]
<div class="warning">No organization found.</div>
[% END %]
</div>`

	approverVars := map[string]interface{}{
		"user": map[string]interface{}{
			"org": map[string]interface{}{
				"name": "Acme Corp",
				"department": map[string]interface{}{
					"name": "Finance",
					"permissions": map[string]interface{}{
						"can_approve":  true,
						"can_view":     true,
						"budget_limit": 50000,
					},
				},
			},
		},
	}

	result := evalTemplate(t, tmpl, approverVars)
	if !strings.Contains(result, `class="action approve"`) {
		t.Error("expected approve action for user with can_approve=true (4 levels deep)")
	}
	if !strings.Contains(result, "Budget limit: $50000") {
		t.Error("expected budget limit at 4th level of dot access")
	}
	if strings.Contains(result, "view-only") || strings.Contains(result, "no-access") {
		t.Error("should not show other permission branches")
	}

	viewerVars := map[string]interface{}{
		"user": map[string]interface{}{
			"org": map[string]interface{}{
				"name": "Acme Corp",
				"department": map[string]interface{}{
					"name": "Marketing",
					"permissions": map[string]interface{}{
						"can_approve": false,
						"can_view":    true,
					},
				},
			},
		},
	}

	result = evalTemplate(t, tmpl, viewerVars)
	if !strings.Contains(result, `class="action view-only"`) {
		t.Error("expected view-only for user with can_view but not can_approve")
	}

	noDeptVars := map[string]interface{}{
		"user": map[string]interface{}{
			"org": map[string]interface{}{
				"name": "Startup Inc",
			},
		},
	}

	result = evalTemplate(t, tmpl, noDeptVars)
	if !strings.Contains(result, "No department assigned.") {
		t.Error("expected 'No department assigned' when department is nil")
	}

	noOrgVars := map[string]interface{}{
		"user": map[string]interface{}{},
	}

	result = evalTemplate(t, tmpl, noOrgVars)
	if !strings.Contains(result, "No organization found.") {
		t.Error("expected 'No organization found' when org is nil")
	}
}

func TestMultiLevelHTMLNestedMenu(t *testing.T) {
	tmpl := `<nav class="menu">
<ul class="level-1">
[% FOREACH item IN menu %]
  <li>
    <a href="[% item.url %]">[% item.label | html %]</a>
    [% IF item.children.size %]
    <ul class="level-2">
    [% FOREACH child IN item.children %]
      <li>
        <a href="[% child.url %]">[% child.label | html %]</a>
        [% IF child.children.size %]
        <ul class="level-3">
        [% FOREACH grandchild IN child.children %]
          <li><a href="[% grandchild.url %]">[% grandchild.label | html %]</a></li>
        [% END %]
        </ul>
        [% END %]
      </li>
    [% END %]
    </ul>
    [% END %]
  </li>
[% END %]
</ul>
</nav>`

	vars := map[string]interface{}{
		"menu": []interface{}{
			map[string]interface{}{
				"label": "Home",
				"url":   "/",
			},
			map[string]interface{}{
				"label": "Products",
				"url":   "/products",
				"children": []interface{}{
					map[string]interface{}{
						"label": "Electronics",
						"url":   "/products/electronics",
						"children": []interface{}{
							map[string]interface{}{"label": "Phones & Tablets", "url": "/products/electronics/phones"},
							map[string]interface{}{"label": "Laptops", "url": "/products/electronics/laptops"},
						},
					},
					map[string]interface{}{
						"label": "Books",
						"url":   "/products/books",
					},
				},
			},
			map[string]interface{}{
				"label": "About",
				"url":   "/about",
			},
		},
	}

	result := evalTemplate(t, tmpl, vars)

	if !strings.Contains(result, `class="level-1"`) {
		t.Error("expected level-1 list")
	}
	if !strings.Contains(result, `class="level-2"`) {
		t.Error("expected level-2 nested list under Products")
	}
	if !strings.Contains(result, `class="level-3"`) {
		t.Error("expected level-3 nested list under Electronics")
	}
	if !strings.Contains(result, "Phones &amp; Tablets") {
		t.Error("expected HTML-escaped label at level 3")
	}
	if !strings.Contains(result, `href="/products/electronics/laptops"`) {
		t.Error("expected deep URL at level 3")
	}
	if strings.Count(result, `class="level-2"`) != 1 {
		t.Errorf("expected exactly 1 level-2 list (only Products has children), got %d", strings.Count(result, `class="level-2"`))
	}
	if strings.Count(result, `class="level-3"`) != 1 {
		t.Errorf("expected exactly 1 level-3 list (only Electronics has children), got %d", strings.Count(result, `class="level-3"`))
	}
}

func TestMultiLevelThreadedComments(t *testing.T) {
	tmpl := `<div class="comments">
[% FOREACH comment IN comments %]
<div class="comment depth-0">
  <strong>[% comment.author | html %]</strong>
  <p>[% comment.text | html %]</p>
  [% IF comment.replies.size %]
  [% FOREACH reply IN comment.replies %]
  <div class="comment depth-1">
    <strong>[% reply.author | html %]</strong>
    <p>[% reply.text | html %]</p>
    [% IF reply.replies.size %]
    [% FOREACH subreply IN reply.replies %]
    <div class="comment depth-2">
      <strong>[% subreply.author | html %]</strong>
      <p>[% subreply.text | html %]</p>
    </div>
    [% END %]
    [% END %]
  </div>
  [% END %]
  [% END %]
</div>
[% END %]
</div>`

	vars := map[string]interface{}{
		"comments": []interface{}{
			map[string]interface{}{
				"author": "Alice",
				"text":   "Great article!",
				"replies": []interface{}{
					map[string]interface{}{
						"author": "Bob <admin>",
						"text":   "Thanks Alice!",
						"replies": []interface{}{
							map[string]interface{}{
								"author": "Alice",
								"text":   "You're welcome & keep writing!",
							},
						},
					},
					map[string]interface{}{
						"author": "Carol",
						"text":   "I agree with Alice.",
					},
				},
			},
			map[string]interface{}{
				"author": "Dave",
				"text":   "Interesting perspective.",
			},
		},
	}

	result := evalTemplate(t, tmpl, vars)

	if strings.Count(result, `class="comment depth-0"`) != 2 {
		t.Errorf("expected 2 top-level comments, got %d", strings.Count(result, `class="comment depth-0"`))
	}
	if strings.Count(result, `class="comment depth-1"`) != 2 {
		t.Errorf("expected 2 depth-1 replies, got %d", strings.Count(result, `class="comment depth-1"`))
	}
	if strings.Count(result, `class="comment depth-2"`) != 1 {
		t.Errorf("expected 1 depth-2 sub-reply, got %d", strings.Count(result, `class="comment depth-2"`))
	}
	if !strings.Contains(result, "Bob &lt;admin&gt;") {
		t.Error("expected HTML-escaped author at depth 1")
	}
	if !strings.Contains(result, "welcome &amp; keep writing!") {
		t.Error("expected HTML-escaped text at depth 2")
	}
	if !strings.Contains(result, "Interesting perspective.") {
		t.Error("expected second top-level comment content")
	}
}

func TestMultiLevelFormWizardSteps(t *testing.T) {
	tmpl := `<div class="wizard">
[% FOREACH step IN steps %]
<fieldset class="step" id="step-[% loop.count %]">
  <legend>Step [% loop.count %]: [% step.title | html %]</legend>
  [% FOREACH field IN step.fields %]
  <div class="field">
    [% SWITCH field.type %]
    [% CASE 'text' %]
    <label>[% field.label | html %]</label>
    <input type="text" name="[% field.name %]" value="[% field.default | html %]">
    [% CASE 'select' %]
    <label>[% field.label | html %]</label>
    <select name="[% field.name %]">
      [% FOREACH group IN field.optgroups %]
      <optgroup label="[% group.label | html %]">
        [% FOREACH opt IN group.options %]
        <option value="[% opt.value %]"[% IF opt.value == field.default %] selected[% END %]>[% opt.label | html %]</option>
        [% END %]
      </optgroup>
      [% END %]
    </select>
    [% CASE 'checkbox' %]
    <label>
      <input type="checkbox" name="[% field.name %]"[% IF field.checked %] checked[% END %]>
      [% field.label | html %]
    </label>
    [% END %]
  </div>
  [% END %]
</fieldset>
[% END %]
</div>`

	vars := map[string]interface{}{
		"steps": []interface{}{
			map[string]interface{}{
				"title": "Personal Info",
				"fields": []interface{}{
					map[string]interface{}{"type": "text", "name": "fullname", "label": "Full Name", "default": ""},
					map[string]interface{}{"type": "text", "name": "email", "label": "Email", "default": "user@example.com"},
				},
			},
			map[string]interface{}{
				"title": "Preferences & Settings",
				"fields": []interface{}{
					map[string]interface{}{
						"type": "select", "name": "country", "label": "Country", "default": "us",
						"optgroups": []interface{}{
							map[string]interface{}{
								"label": "North America",
								"options": []interface{}{
									map[string]interface{}{"value": "us", "label": "United States"},
									map[string]interface{}{"value": "ca", "label": "Canada"},
								},
							},
							map[string]interface{}{
								"label": "Europe",
								"options": []interface{}{
									map[string]interface{}{"value": "uk", "label": "United Kingdom"},
									map[string]interface{}{"value": "de", "label": "Germany"},
								},
							},
						},
					},
					map[string]interface{}{"type": "checkbox", "name": "newsletter", "label": "Subscribe to newsletter", "checked": true},
				},
			},
			map[string]interface{}{
				"title": "Review",
				"fields": []interface{}{
					map[string]interface{}{"type": "checkbox", "name": "agree", "label": "I agree to terms & conditions", "checked": false},
				},
			},
		},
	}

	result := evalTemplate(t, tmpl, vars)

	if strings.Count(result, `class="step"`) != 3 {
		t.Errorf("expected 3 wizard steps, got %d", strings.Count(result, `class="step"`))
	}
	if !strings.Contains(result, "Step 1: Personal Info") {
		t.Error("expected step 1 title with loop.count")
	}
	if !strings.Contains(result, "Preferences &amp; Settings") {
		t.Error("expected HTML-escaped step title")
	}
	if !strings.Contains(result, `<optgroup label="North America">`) {
		t.Error("expected optgroup at nesting level 4 (step > field > optgroup)")
	}
	if !strings.Contains(result, `<option value="us" selected>United States</option>`) {
		t.Error("expected selected option at nesting level 5 (step > field > optgroup > option)")
	}
	if strings.Count(result, " selected") != 1 {
		t.Errorf("expected exactly 1 selected option, got %d", strings.Count(result, " selected"))
	}
	if !strings.Contains(result, " checked") {
		t.Error("expected checked attribute on newsletter checkbox")
	}
	if !strings.Contains(result, "terms &amp; conditions") {
		t.Error("expected HTML-escaped checkbox label")
	}
	if strings.Count(result, "<optgroup") != 2 {
		t.Errorf("expected 2 optgroups, got %d", strings.Count(result, "<optgroup"))
	}
}

func TestMultiLevelDashboardWidgets(t *testing.T) {
	tmpl := `<div class="dashboard">
[% FOREACH section IN sections %]
<section class="section">
  <h1>[% section.title | html %]</h1>
  [% FOREACH row IN section.rows %]
  <div class="row">
    [% FOREACH widget IN row.widgets %]
    <div class="widget widget-[% widget.type %]">
      <h3>[% widget.title | html %]</h3>
      [% SWITCH widget.type %]
      [% CASE 'stat' %]
      <p class="stat-value">[% widget.data.value %]</p>
      <p class="stat-label">[% widget.data.label | html %]</p>
      [% CASE 'list' %]
      <ul>
        [% FOREACH item IN widget.data.items %]
        <li>[% item.name | html %]: [% item.value %][% IF item.trend == 'up' %] &#8593;[% ELSIF item.trend == 'down' %] &#8595;[% END %]</li>
        [% END %]
      </ul>
      [% CASE 'chart' %]
      <div class="chart" data-type="[% widget.data.chart_type %]">
        [% FOREACH point IN widget.data.points %]
        <span data-x="[% point.x %]" data-y="[% point.y %]"></span>
        [% END %]
      </div>
      [% CASE %]
      <p>Unknown widget type</p>
      [% END %]
    </div>
    [% END %]
  </div>
  [% END %]
</section>
[% END %]
</div>`

	vars := map[string]interface{}{
		"sections": []interface{}{
			map[string]interface{}{
				"title": "Overview",
				"rows": []interface{}{
					map[string]interface{}{
						"widgets": []interface{}{
							map[string]interface{}{
								"type":  "stat",
								"title": "Total Users",
								"data": map[string]interface{}{
									"value": 1250,
									"label": "Active & Registered",
								},
							},
							map[string]interface{}{
								"type":  "stat",
								"title": "Revenue",
								"data": map[string]interface{}{
									"value": 84500,
									"label": "This Month",
								},
							},
						},
					},
					map[string]interface{}{
						"widgets": []interface{}{
							map[string]interface{}{
								"type":  "list",
								"title": "Top Products",
								"data": map[string]interface{}{
									"items": []interface{}{
										map[string]interface{}{"name": "Widget Pro", "value": 320, "trend": "up"},
										map[string]interface{}{"name": "Gadget <Basic>", "value": 280, "trend": "down"},
										map[string]interface{}{"name": "Tool Kit", "value": 150, "trend": ""},
									},
								},
							},
						},
					},
				},
			},
			map[string]interface{}{
				"title": "Analytics",
				"rows": []interface{}{
					map[string]interface{}{
						"widgets": []interface{}{
							map[string]interface{}{
								"type":  "chart",
								"title": "Traffic",
								"data": map[string]interface{}{
									"chart_type": "line",
									"points": []interface{}{
										map[string]interface{}{"x": "Jan", "y": 100},
										map[string]interface{}{"x": "Feb", "y": 150},
										map[string]interface{}{"x": "Mar", "y": 130},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	result := evalTemplate(t, tmpl, vars)

	if strings.Count(result, `class="section"`) != 2 {
		t.Errorf("expected 2 sections, got %d", strings.Count(result, `class="section"`))
	}
	if strings.Count(result, `class="row"`) != 3 {
		t.Errorf("expected 3 rows total, got %d", strings.Count(result, `class="row"`))
	}
	if !strings.Contains(result, `class="widget widget-stat"`) {
		t.Error("expected stat widget type in class")
	}
	if !strings.Contains(result, `class="widget widget-list"`) {
		t.Error("expected list widget type in class")
	}
	if !strings.Contains(result, `class="widget widget-chart"`) {
		t.Error("expected chart widget type in class")
	}
	if !strings.Contains(result, `class="stat-value">1250</p>`) {
		t.Error("expected stat value at level 4 (section > row > widget > data)")
	}
	if !strings.Contains(result, "Active &amp; Registered") {
		t.Error("expected HTML-escaped stat label at level 4")
	}
	if !strings.Contains(result, "Gadget &lt;Basic&gt;") {
		t.Error("expected HTML-escaped list item name at level 5 (section > row > widget > items > item)")
	}
	if !strings.Contains(result, "&#8593;") {
		t.Error("expected up arrow for trend=up at level 5")
	}
	if !strings.Contains(result, "&#8595;") {
		t.Error("expected down arrow for trend=down at level 5")
	}
	if !strings.Contains(result, `data-type="line"`) {
		t.Error("expected chart type attribute at level 4")
	}
	if strings.Count(result, "data-x=") != 3 {
		t.Errorf("expected 3 chart data points, got %d", strings.Count(result, "data-x="))
	}
}
