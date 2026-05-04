const servicesData = [
  {
    title: "Традиционные похороны",
    description: "Полный комплекс услуг: прощание, церемония и погребение. Включает использование помещения, транспорт и работу персонала.",
    price: "450 000 ₽"
  },
  {
    title: "Кремация",
    description: "Достойная кремация с возможностью проведения поминальной церемонии. Включает контейнер, урну и процесс кремации.",
    price: "280 000 ₽"
  },
  {
    title: "Поминальная церемония",
    description: "Индивидуальная церемония памяти без присутствия гроба. Включает помещение, персонал и координацию.",
    price: "150 000 ₽"
  },
  {
    title: "Церемония на кладбище",
    description: "Камерная церемония на месте захоронения с профессиональной организацией.",
    price: "120 000 ₽"
  },
  {
    title: "Простое погребение",
    description: "Погребение без церемонии. Включает базовые услуги, транспорт и оформление документов.",
    price: "220 000 ₽"
  },
  {
    title: "Прощание",
    description: "Частное или публичное прощание с использованием помещения и профессиональной подготовкой.",
    price: "80 000 ₽"
  },
  {
    title: "Транспортировка",
    description: "Профессиональная транспортировка в пределах города.",
    price: "35 000 ₽"
  },
  {
    title: "Бальзамирование",
    description: "Профессиональная подготовка и консервация.",
    price: "65 000 ₽"
  },
  {
    title: "Услуги некролога",
    description: "Профессиональное написание и размещение некролога.",
    price: "20 000 ₽"
  }
];

export function Services() {
  return (
    <div className="max-w-7xl mx-auto px-6 py-12">
      <div className="mb-12 text-center">
        <h1 className="text-4xl mb-4 text-foreground">Наши услуги</h1>
        <p className="text-lg text-muted-foreground max-w-2xl mx-auto">
          Профессиональные ритуальные услуги для достойного прощания с вашими близкими
        </p>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {servicesData.map((service) => (
          <div
            key={service.title}
            className="bg-card border border-border rounded-lg p-6 hover:shadow-md transition-shadow"
          >
            <h3 className="text-xl mb-3 text-foreground">{service.title}</h3>
            <p className="text-muted-foreground mb-4 min-h-[72px]">
              {service.description}
            </p>
            <div className="flex items-center justify-between pt-4 border-t border-border">
              <span className="text-xl text-primary">{service.price}</span>
            </div>
          </div>
        ))}
      </div>

      <div className="mt-12 bg-secondary border border-border rounded-lg p-8 text-center">
        <h2 className="text-2xl mb-3 text-foreground">Нужна помощь в выборе?</h2>
        <p className="text-muted-foreground mb-6">
          Наши опытные сотрудники помогут вам подобрать необходимые услуги
        </p>
        <a
          href="tel:+71234567890"
          className="inline-block bg-primary text-primary-foreground px-8 py-3 rounded-lg hover:opacity-90 transition-opacity"
        >
          Позвонить для консультации
        </a>
      </div>
    </div>
  );
}
