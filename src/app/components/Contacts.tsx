export function Contacts() {
  return (
    <div className="max-w-4xl mx-auto px-6 py-12">
      <div className="mb-12 text-center">
        <h1 className="text-4xl mb-4 text-foreground">Контакты</h1>
        <p className="text-lg text-muted-foreground">
          Мы готовы помочь вам круглосуточно в этот сложный период
        </p>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
        {/* Contact Information */}
        <div className="bg-card border border-border rounded-lg p-8">
          <h2 className="text-2xl mb-6 text-foreground">Свяжитесь с нами</h2>
          <div className="space-y-6">
            <div>
              <h3 className="mb-2 text-foreground">Телефон</h3>
              <a
                href="tel:+71234567890"
                className="text-primary hover:underline text-lg"
              >
                +7 (123) 456-78-90
              </a>
              <p className="text-sm text-muted-foreground mt-1">
                Работаем круглосуточно
              </p>
            </div>

            <div>
              <h3 className="mb-2 text-foreground">Email</h3>
              <a
                href="mailto:info@ritual.ru"
                className="text-primary hover:underline"
              >
                info@ritual.ru
              </a>
            </div>

            <div>
              <h3 className="mb-2 text-foreground">Адрес</h3>
              <p className="text-muted-foreground">
                г. Москва
                <br />
                ул. Примерная, д. 123
              </p>
            </div>

            <div>
              <h3 className="mb-2 text-foreground">Часы работы</h3>
              <p className="text-muted-foreground">
                Офис: Пн-Пт, 9:00 - 18:00
                <br />
                Экстренная помощь: круглосуточно
              </p>
            </div>
          </div>
        </div>

        {/* Contact Form */}
        <div className="bg-card border border-border rounded-lg p-8">
          <h2 className="text-2xl mb-6 text-foreground">Отправить сообщение</h2>
          <form className="space-y-6">
            <div>
              <label htmlFor="contact-name" className="block mb-2 text-foreground">
                Ваше имя *
              </label>
              <input
                id="contact-name"
                type="text"
                required
                className="w-full px-4 py-3 bg-input-background border border-border rounded-lg focus:outline-none focus:ring-2 focus:ring-ring"
                placeholder="Иванов Иван"
              />
            </div>

            <div>
              <label htmlFor="contact-email" className="block mb-2 text-foreground">
                Email *
              </label>
              <input
                id="contact-email"
                type="email"
                required
                className="w-full px-4 py-3 bg-input-background border border-border rounded-lg focus:outline-none focus:ring-2 focus:ring-ring"
                placeholder="ivan@example.com"
              />
            </div>

            <div>
              <label htmlFor="contact-phone" className="block mb-2 text-foreground">
                Номер телефона
              </label>
              <input
                id="contact-phone"
                type="tel"
                className="w-full px-4 py-3 bg-input-background border border-border rounded-lg focus:outline-none focus:ring-2 focus:ring-ring"
                placeholder="+7 (123) 456-78-90"
              />
            </div>

            <div>
              <label htmlFor="contact-message" className="block mb-2 text-foreground">
                Сообщение *
              </label>
              <textarea
                id="contact-message"
                required
                rows={4}
                className="w-full px-4 py-3 bg-input-background border border-border rounded-lg focus:outline-none focus:ring-2 focus:ring-ring resize-none"
                placeholder="Чем мы можем вам помочь?"
              />
            </div>

            <button
              type="submit"
              className="w-full bg-primary text-primary-foreground py-3 rounded-lg hover:opacity-90 transition-opacity"
            >
              Отправить сообщение
            </button>
          </form>
        </div>
      </div>

      {/* Map Placeholder */}
      <div className="mt-12 bg-card border border-border rounded-lg p-8">
        <h2 className="text-2xl mb-6 text-foreground text-center">Наше местоположение</h2>
        <div className="bg-muted rounded-lg h-64 flex items-center justify-center">
          <p className="text-muted-foreground">Карта - ул. Примерная, д. 123</p>
        </div>
      </div>
    </div>
  );
}
