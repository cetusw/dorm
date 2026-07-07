import './App.css'

function App() {
  return (
      <main className="app-page">
        <section className="app-card">
          <h1>Сервис уборок</h1>
          <p>Первая страница приложения для жителей.</p>

          <div className="app-actions">
            <button type="button">Мои задачи</button>
            <button type="button">Текущая неделя</button>
          </div>
        </section>
      </main>
  )
}

export default App