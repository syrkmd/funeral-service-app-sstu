import { Link } from "react-router";

const services = [
  {
    title: "Традиционные похороны",
    description: "Полный комплекс услуг: прощание, церемония и погребение",
    price: "450 000 ₽"
  },
  {
    title: "Кремация",
    description: "Достойная кремация с возможностью проведения поминальной церемонии",
    price: "280 000 ₽"
  },
  {
    title: "Поминальная церемония",
    description: "Индивидуальная церемония памяти",
    price: "150 000 ₽"
  }
];

const products = [
  {
    title: "Гроб дубовый",
    description: "Массив дуба премиум-класса с атласной отделкой",
    price: "320 000 ₽"
  },
  {
    title: "Траурная композиция",
    description: "Свежие сезонные цветы, индивидуальный дизайн",
    price: "25 000 ₽"
  },
  {
    title: "Урна",
    description: "Латунная урна с возможностью гравировки",
    price: "45 000 ₽"
  }
];

export function Home() {
  return (
    <div>
      {/* Hero Section */}
      <section className="bg-gradient-to-b from-secondary to-background py-20">
        <div className="max-w-4xl mx-auto px-6 text-center">
          <h1 className="text-4xl mb-4 text-foreground">
            Провожаем с уважением и заботой
          </h1>
          <p className="text-lg text-muted-foreground mb-8 max-w-2xl mx-auto">
            Мы понимаем, насколько это сложный период. Позвольте нам помочь вам в организации достойного прощания.
          </p>
          <Link
            to="/order"
            className="inline-block bg-primary text-primary-foreground px-8 py-3 rounded-lg hover:opacity-90 transition-opacity"
          >
            Оформить услугу
          </Link>
        </div>
      </section>

      {/* Services Preview */}
      <section className="max-w-7xl mx-auto px-6 py-16">
        <h2 className="text-3xl mb-8 text-center text-foreground">Наши услуги</h2>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
          {services.map((service) => (
            <div
              key={service.title}
              className="bg-card border border-border rounded-lg p-6 hover:shadow-sm transition-shadow"
            >
              <h3 className="text-xl mb-3 text-foreground">{service.title}</h3>
              <p className="text-muted-foreground mb-4 min-h-[48px]">
                {service.description}
              </p>
              <p className="text-primary">{service.price}</p>
            </div>
          ))}
        </div>
        <div className="text-center">
          <Link
            to="/services"
            className="text-primary hover:underline"
          >
            Все услуги →
          </Link>
        </div>
      </section>

      {/* Products Preview */}
      <section className="bg-secondary py-16">
        <div className="max-w-7xl mx-auto px-6">
          <h2 className="text-3xl mb-8 text-center text-foreground">Товары</h2>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
            {products.map((product) => (
              <div
                key={product.title}
                className="bg-card border border-border rounded-lg p-6 hover:shadow-sm transition-shadow"
              >
                <div className="bg-muted rounded-md h-32 mb-4 flex items-center justify-center">
                  <span className="text-muted-foreground text-sm">{product.title}</span>
                </div>
                <h3 className="mb-2 text-foreground">{product.title}</h3>
                <p className="text-sm text-muted-foreground mb-3">
                  {product.description}
                </p>
                <p className="text-primary">{product.price}</p>
              </div>
            ))}
          </div>
          <div className="text-center">
            <Link
              to="/catalog"
              className="text-primary hover:underline"
            >
              Весь каталог →
            </Link>
          </div>
        </div>
      </section>

      {/* Contact Section */}
      <section className="max-w-4xl mx-auto px-6 py-16 text-center">
        <h2 className="text-3xl mb-4 text-foreground">Мы готовы помочь</h2>
        <p className="text-muted-foreground mb-6">
          Работаем круглосуточно, без выходных
        </p>
        <div className="flex flex-col sm:flex-row gap-4 justify-center">
          <a
            href="tel:+1234567890"
            className="bg-primary text-primary-foreground px-6 py-3 rounded-lg hover:opacity-90 transition-opacity"
          >
            Позвонить: +7 (123) 456-78-90
          </a>
          <Link
            to="/contacts"
            className="bg-secondary text-secondary-foreground border border-border px-6 py-3 rounded-lg hover:bg-accent transition-colors"
          >
            Контактная информация
          </Link>
        </div>
      </section>
    </div>
  );
}
