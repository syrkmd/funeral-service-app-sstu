import { useEffect, useState } from "react";
import { useNavigate } from "react-router";
import {
  getCemeteryPlots,
  getCemeterySections,
  type CemeteryPlot,
  type CemeterySection,
} from "../../api/cemetery.api";
import { CatalogItemImage } from "./CatalogItemImage";
import { useOrdersStore } from "../store/ordersStore";
import { normalizePhone } from "../utils/phoneUtils";

const availableServices = [
  { id: "traditional", name: "Традиционные похороны", price: 450000 },
  { id: "cremation", name: "Кремация", price: 280000 },
  { id: "memorial", name: "Поминальная церемония", price: 150000 },
  { id: "graveside", name: "Церемония на кладбище", price: 120000 },
  { id: "viewing", name: "Прощание", price: 80000 },
  { id: "embalming", name: "Бальзамирование", price: 65000 },
  { id: "transportation", name: "Транспортировка", price: 35000 },
];

const availableProducts = [
  { id: "oak-casket", name: "Гроб дубовый премиум", price: 320000 },
  { id: "mahogany-casket", name: "Гроб из красного дерева", price: 410000 },
  { id: "pine-casket", name: "Гроб сосновый", price: 180000 },
  { id: "brass-urn", name: "Урна латунная", price: 45000 },
  { id: "ceramic-urn", name: "Урна керамическая", price: 32000 },
  { id: "casket-spray", name: "Траурная композиция на гроб", price: 45000 },
  { id: "wreath", name: "Венок", price: 28000 },
  { id: "guest-book", name: "Книга памяти", price: 7500 },
];

type FormData = {
  clientName: string;
  clientPhone: string;
  clientEmail: string;
  deceasedName: string;
  dateOfBirth: string;
  dateOfDeath: string;
  selectedServices: string[];
  selectedProducts: string[];
  serviceType: string;
  serviceDate: string;
  serviceTime: string;
  serviceAddress: string;
  cemetery: string;
  cemeteryOther: string;
  cemeteryNotes: string;
  selectedSectionName: string;
  selectedPlotId: number | null;
  selectedPlotLabel: string;
  paymentMethod: string;
  comments: string;
};

type PaymentDetails = {
  cardNumber: string;
  expiryDate: string;
  cvv: string;
};

type FormErrors = {
  clientName?: string;
  clientPhone?: string;
  clientEmail?: string;
  deceasedName?: string;
  dateOfBirth?: string;
  dateOfDeath?: string;
  serviceDate?: string;
  cemeteryPlot?: string;
};

const PERSON_NAME_PATTERN = /^[\p{L}]+(?:[ '\-][\p{L}]+)*$/u;

export function OrderForm() {
  const createOrder = useOrdersStore((state) => state.createOrder);
  const payOrder = useOrdersStore((state) => state.payOrder);
  const [currentStep, setCurrentStep] = useState(1);
  const [orderId, setOrderId] = useState<string | null>(null);
  const [paymentResult, setPaymentResult] = useState<"paid" | "deferred" | "failed" | null>(null);
  const [paymentError, setPaymentError] = useState("");
  const [submissionError, setSubmissionError] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [errors, setErrors] = useState<FormErrors>({});
  const [completedSteps, setCompletedSteps] = useState<Set<number>>(new Set());
  const [cemeterySections, setCemeterySections] = useState<CemeterySection[]>([]);
  const [cemeteryPlots, setCemeteryPlots] = useState<CemeteryPlot[]>([]);
  const [isLoadingPlots, setIsLoadingPlots] = useState(false);
  const [plotsError, setPlotsError] = useState<string | null>(null);

  const [formData, setFormData] = useState<FormData>({
    clientName: "",
    clientPhone: "",
    clientEmail: "",
    deceasedName: "",
    dateOfBirth: "",
    dateOfDeath: "",
    selectedServices: [],
    selectedProducts: [],
    serviceType: "",
    serviceDate: "",
    serviceTime: "",
    serviceAddress: "",
    cemetery: "",
    cemeteryOther: "",
    cemeteryNotes: "",
    selectedSectionName: "",
    selectedPlotId: null,
    selectedPlotLabel: "",
    paymentMethod: "later",
    comments: "",
  });
  const [paymentDetails, setPaymentDetails] = useState<PaymentDetails>({
    cardNumber: "",
    expiryDate: "",
    cvv: "",
  });

  const steps = [
    "Данные клиента",
    "Данные умершего",
    "Услуги",
    "Товары",
    "Детали",
    "Проверка"
  ];

  useEffect(() => {
    if (currentStep !== 5 || formData.serviceType !== "burial" || cemeterySections.length > 0) {
      return;
    }

    getCemeterySections()
      .then(setCemeterySections)
      .catch(() => setPlotsError("Не удалось загрузить секции кладбища"));
  }, [currentStep, formData.serviceType, cemeterySections.length]);

  useEffect(() => {
    if (
      currentStep !== 5 ||
      formData.serviceType !== "burial" ||
      !formData.selectedSectionName
    ) {
      return;
    }

    let cancelled = false;
    setIsLoadingPlots(true);
    setPlotsError(null);

    getCemeteryPlots(formData.selectedSectionName)
      .then((plots) => {
        if (!cancelled) {
          setCemeteryPlots(plots);
        }
      })
      .catch(() => {
        if (!cancelled) {
          setCemeteryPlots([]);
          setPlotsError("Не удалось загрузить места захоронения");
        }
      })
      .finally(() => {
        if (!cancelled) {
          setIsLoadingPlots(false);
        }
      });

    return () => {
      cancelled = true;
    };
  }, [currentStep, formData.serviceType, formData.selectedSectionName]);

  useEffect(() => {
    if (formData.serviceType === "burial" || formData.selectedPlotId === null) {
      return;
    }

    setFormData(prev => ({
      ...prev,
      selectedSectionName: "",
      selectedPlotId: null,
      selectedPlotLabel: "",
    }));
  }, [formData.serviceType, formData.selectedPlotId]);

  const updateFormData = (field: keyof FormData, value: any) => {
    setFormData(prev => ({ ...prev, [field]: value }));
  };

  const updatePaymentDetails = (field: keyof PaymentDetails, value: string) => {
    setPaymentDetails(prev => ({ ...prev, [field]: value }));
  };

  const handleCardNumberChange = (value: string) => {
    updatePaymentDetails("cardNumber", value.replace(/\D/g, "").slice(0, 16));
  };

  const handleExpiryChange = (value: string) => {
    const digits = value.replace(/\D/g, "").slice(0, 4);
    updatePaymentDetails(
      "expiryDate",
      digits.length > 2 ? `${digits.slice(0, 2)}/${digits.slice(2)}` : digits,
    );
  };

  const isCardDetailsValid =
    /^\d{16}$/.test(paymentDetails.cardNumber) &&
    /^(0[1-9]|1[0-2])\/\d{2}$/.test(paymentDetails.expiryDate) &&
    /^\d{3}$/.test(paymentDetails.cvv);

  const toggleService = (id: string) => {
    setFormData(prev => ({
      ...prev,
      selectedServices: prev.selectedServices.includes(id)
        ? prev.selectedServices.filter(s => s !== id)
        : [...prev.selectedServices, id]
    }));
  };

  const removeService = (id: string) => {
    if (formData.selectedServices.length <= 1) {
      return;
    }
    setFormData(prev => ({
      ...prev,
      selectedServices: prev.selectedServices.filter(s => s !== id)
    }));
  };

  const toggleProduct = (id: string) => {
    setFormData(prev => ({
      ...prev,
      selectedProducts: prev.selectedProducts.includes(id)
        ? prev.selectedProducts.filter(p => p !== id)
        : [...prev.selectedProducts, id]
    }));
  };

  const selectPlot = (plot: CemeteryPlot) => {
    if (!plot.available) {
      return;
    }

    setFormData(prev => ({
      ...prev,
      selectedPlotId: plot.id,
      selectedPlotLabel: plot.label,
    }));
  };

  const selectSection = (sectionName: string) => {
    setCemeteryPlots([]);
    setPlotsError(null);
    setFormData(prev => ({
      ...prev,
      selectedSectionName: sectionName,
      selectedPlotId: null,
      selectedPlotLabel: "",
    }));
  };

  const removeProduct = (id: string) => {
    setFormData(prev => ({
      ...prev,
      selectedProducts: prev.selectedProducts.filter(p => p !== id)
    }));
  };

  const validateStep = (step: number): boolean => {
    const newErrors: FormErrors = {};

    if (step === 1) {
      if (!formData.clientName.trim()) {
        newErrors.clientName = "Обязательное поле";
      } else if (!PERSON_NAME_PATTERN.test(formData.clientName.trim())) {
        newErrors.clientName = "Имя может содержать только буквы, пробелы и дефисы";
      }
      if (!formData.clientPhone.trim()) {
        newErrors.clientPhone = "Обязательное поле";
      } else {
        const normalizedPhone = normalizePhone(formData.clientPhone);
        if (!/^\+?\d{11,15}$/.test(normalizedPhone)) {
          newErrors.clientPhone = "Введите корректный номер телефона";
        }
      }
      if (formData.clientEmail && !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(formData.clientEmail)) {
        newErrors.clientEmail = "Некорректный email";
      }
    }

    if (step === 2) {
      if (!formData.deceasedName.trim()) {
        newErrors.deceasedName = "Обязательное поле";
      } else if (!PERSON_NAME_PATTERN.test(formData.deceasedName.trim())) {
        newErrors.deceasedName = "Имя может содержать только буквы, пробелы и дефисы";
      }
      if (!formData.dateOfDeath) {
        newErrors.dateOfDeath = "Обязательное поле";
      } else if (
        formData.dateOfBirth &&
        formData.dateOfDeath <= formData.dateOfBirth
      ) {
        newErrors.dateOfDeath = "Дата смерти должна быть позже даты рождения";
      }
    }

    if (step === 5 && formData.serviceType === "burial") {
      if (!formData.serviceDate) {
        newErrors.serviceDate = "Укажите дату церемонии для бронирования участка";
      }
      if (formData.selectedPlotId === null) {
        newErrors.cemeteryPlot = "Выберите свободное место захоронения";
      }
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const calculateTotal = () => {
    const servicesTotal = availableServices
      .filter(s => formData.selectedServices.includes(s.id))
      .reduce((sum, s) => sum + s.price, 0);

    const productsTotal = availableProducts
      .filter(p => formData.selectedProducts.includes(p.id))
      .reduce((sum, p) => sum + p.price, 0);

    return servicesTotal + productsTotal;
  };

  const nextStep = () => {
    if (validateStep(currentStep) && currentStep < 6) {
      setCompletedSteps(prev => new Set(prev).add(currentStep));
      setCurrentStep(currentStep + 1);
      window.scrollTo({ top: 0, behavior: 'smooth' });
    }
  };

  const prevStep = () => {
    if (currentStep > 1) {
      setCurrentStep(currentStep - 1);
      setErrors({});
      window.scrollTo({ top: 0, behavior: 'smooth' });
    }
  };

  const jumpToStep = (step: number) => {
    if (step <= currentStep || completedSteps.has(step - 1)) {
      setCurrentStep(step);
      setErrors({});
      window.scrollTo({ top: 0, behavior: 'smooth' });
    }
  };

  const canNavigateToStep = (step: number) => {
    return step < currentStep || completedSteps.has(step - 1);
  };

  const handleSubmit = async () => {
    if (isSubmitting) {
      return;
    }

    setIsSubmitting(true);
    setPaymentError("");
    setSubmissionError("");

    const newOrderData = {
      date: new Date().toISOString().split('T')[0],
      client: {
        name: formData.clientName,
        phone: normalizePhone(formData.clientPhone),
        email: formData.clientEmail,
      },
      deceased: {
        name: formData.deceasedName,
        dateOfBirth: formData.dateOfBirth,
        dateOfDeath: formData.dateOfDeath,
      },
      serviceDate: formData.serviceDate || null,
      serviceTime: formData.serviceTime || null,
      serviceAddress: formData.serviceAddress || null,
      cemetery: formData.cemetery === "Другое (указать)"
        ? formData.cemeteryOther
        : formData.cemetery || null,
      cemeteryNotes: formData.cemeteryNotes || null,
      cemeteryPlotId: formData.selectedPlotId,
      cemeteryPlotCode: formData.selectedPlotLabel || null,
      services: formData.selectedServices.map(serviceId => {
        const service = availableServices.find(s => s.id === serviceId);
        return {
          name: service?.name || "",
          price: service?.price || 0,
        };
      }),
      products: formData.selectedProducts.map(productId => {
        const product = availableProducts.find(p => p.id === productId);
        return {
          name: product?.name || "",
          price: product?.price || 0,
        };
      }),
      documents: [],
    };

    try {
      const newOrder = await createOrder(newOrderData);

      if (formData.paymentMethod === "credit") {
        try {
          await payOrder(newOrder.id, paymentDetails);
          setPaymentResult("paid");
        } catch (error) {
          setPaymentResult("failed");
          setPaymentError(
            error instanceof Error
              ? error.message
              : "Заказ создан, но выполнить оплату не удалось",
          );
        }
      } else {
        setPaymentResult("deferred");
      }

      setOrderId(newOrder.id);
      setCurrentStep(7);
      window.scrollTo({ top: 0, behavior: 'smooth' });
    } catch (error) {
      setSubmissionError(
        error instanceof Error
          ? error.message
          : "Не удалось оформить заказ. Проверьте данные и повторите попытку.",
      );
    } finally {
      setIsSubmitting(false);
    }
  };

  const canProceed = () => {
    switch (currentStep) {
      case 1:
        return formData.clientName.trim() &&
               PERSON_NAME_PATTERN.test(formData.clientName.trim()) &&
               formData.clientPhone.trim() &&
               /^[\d\s\-\(\)\+]+$/.test(formData.clientPhone) &&
               (!formData.clientEmail || /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(formData.clientEmail));
      case 2:
        return formData.deceasedName.trim() &&
               PERSON_NAME_PATTERN.test(formData.deceasedName.trim()) &&
               formData.dateOfDeath &&
               (!formData.dateOfBirth || formData.dateOfDeath > formData.dateOfBirth);
      case 3:
        return formData.selectedServices.length > 0;
      case 4:
        return true;
      case 5:
        return (
          (formData.serviceType !== "burial" ||
            (Boolean(formData.serviceDate) && formData.selectedPlotId !== null)) &&
          (formData.paymentMethod !== "credit" || isCardDetailsValid)
        );
      case 6:
        return formData.selectedServices.length > 0;
      default:
        return false;
    }
  };

  if (orderId) {
    return <OrderConfirmation
      orderId={orderId}
      paymentResult={paymentResult}
      paymentError={paymentError}
      onNewOrder={() => {
      setOrderId(null);
      setPaymentResult(null);
      setPaymentError("");
      setSubmissionError("");
      setCurrentStep(1);
      setFormData({
        clientName: "",
        clientPhone: "",
        clientEmail: "",
        deceasedName: "",
        dateOfBirth: "",
        dateOfDeath: "",
        selectedServices: [],
        selectedProducts: [],
        serviceType: "",
        serviceDate: "",
        serviceTime: "",
        serviceAddress: "",
        cemetery: "",
        cemeteryOther: "",
        cemeteryNotes: "",
        selectedSectionName: "",
        selectedPlotId: null,
        selectedPlotLabel: "",
        paymentMethod: "later",
        comments: "",
      });
      setPaymentDetails({
        cardNumber: "",
        expiryDate: "",
        cvv: "",
      });
      setErrors({});
      setCompletedSteps(new Set());
      setCemeteryPlots([]);
    }}
    />;
  }

  return (
    <div className="min-h-screen bg-background py-12">
      <div className="max-w-7xl mx-auto px-6">
        <div className="mb-8 text-center">
          <h1 className="text-4xl mb-4 text-foreground">Оформление услуги</h1>
          <p className="text-lg text-muted-foreground">
            Шаг {currentStep} из 6: {steps[currentStep - 1]}
          </p>
        </div>

        {/* Progress Bar */}
        <div className="max-w-4xl mx-auto mb-12">
          <div className="flex items-center justify-between mb-4">
            {steps.map((step, index) => {
              const stepNumber = index + 1;
              const isClickable = canNavigateToStep(stepNumber);
              const isCurrent = stepNumber === currentStep;
              const isCompleted = stepNumber < currentStep;

              return (
                <div key={step} className="flex-1 flex items-center">
                  <div className="flex flex-col items-center flex-1">
                    <button
                      onClick={() => jumpToStep(stepNumber)}
                      disabled={!isClickable}
                      className={`w-10 h-10 rounded-full flex items-center justify-center text-sm transition-all ${
                        isCurrent
                          ? "bg-primary text-primary-foreground"
                          : isCompleted
                          ? "bg-primary/70 text-primary-foreground hover:bg-primary/80 cursor-pointer"
                          : "bg-secondary text-muted-foreground cursor-not-allowed"
                      } ${isClickable && !isCurrent ? "hover:scale-105" : ""}`}
                      title={isClickable ? `Перейти к: ${step}` : step}
                    >
                      {stepNumber}
                    </button>
                    <div className="text-xs mt-2 text-center text-muted-foreground hidden md:block">
                      {step}
                    </div>
                  </div>
                  {index < steps.length - 1 && (
                    <div
                      className={`h-1 flex-1 mx-2 rounded transition-colors ${
                        stepNumber < currentStep ? "bg-primary/70" : "bg-secondary"
                      }`}
                    />
                  )}
                </div>
              );
            })}
          </div>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
          {/* Main Form */}
          <div className="lg:col-span-2">
            {currentStep === 1 && (
              <Step1ClientInfo formData={formData} updateFormData={updateFormData} errors={errors} />
            )}
            {currentStep === 2 && (
              <Step2DeceasedInfo formData={formData} updateFormData={updateFormData} errors={errors} />
            )}
            {currentStep === 3 && (
              <Step3Services
                selectedServices={formData.selectedServices}
                toggleService={toggleService}
              />
            )}
            {currentStep === 4 && (
              <Step4Products
                selectedProducts={formData.selectedProducts}
                toggleProduct={toggleProduct}
              />
            )}
            {currentStep === 5 && (
              <Step5Details
                formData={formData}
                updateFormData={updateFormData}
                sections={cemeterySections}
                plots={cemeteryPlots}
                selectedPlotId={formData.selectedPlotId}
                isLoadingPlots={isLoadingPlots}
                plotsError={plotsError}
                selectSection={selectSection}
                selectPlot={selectPlot}
                paymentDetails={paymentDetails}
                handleCardNumberChange={handleCardNumberChange}
                handleExpiryChange={handleExpiryChange}
                updatePaymentDetails={updatePaymentDetails}
                errors={errors}
              />
            )}
            {currentStep === 6 && (
              <Step6Review formData={formData} jumpToStep={jumpToStep} />
            )}

            {/* Navigation Buttons */}
            {submissionError && (
              <div className="mt-6 bg-destructive/10 border border-destructive/40 rounded-lg p-4 text-destructive">
                {submissionError}
              </div>
            )}
            <div className="flex gap-4 mt-8">
              {currentStep > 1 && (
                <button
                  onClick={prevStep}
                  className="flex-1 bg-secondary text-secondary-foreground py-3 px-6 rounded-lg hover:bg-accent transition-colors"
                >
                  ← Назад
                </button>
              )}
              {currentStep < 6 ? (
                <button
                  onClick={nextStep}
                  disabled={!canProceed()}
                  className="flex-1 bg-primary text-primary-foreground py-3 px-6 rounded-lg transition-opacity disabled:opacity-40 disabled:cursor-not-allowed disabled:hover:opacity-40 hover:opacity-90"
                >
                  Далее →
                </button>
              ) : (
                <button
                  onClick={handleSubmit}
                  disabled={!canProceed() || isSubmitting}
                  className="flex-1 bg-primary text-primary-foreground py-3 px-6 rounded-lg transition-opacity disabled:opacity-40 disabled:cursor-not-allowed disabled:hover:opacity-40 hover:opacity-90"
                >
                  {isSubmitting
                    ? "Оформление..."
                    : formData.paymentMethod === "credit"
                      ? "Оформить и оплатить"
                      : "Отправить заказ"}
                </button>
              )}
            </div>
          </div>

          {/* Order Summary Sidebar */}
          <div className="lg:col-span-1">
            <OrderSummary
              selectedServices={formData.selectedServices}
              selectedProducts={formData.selectedProducts}
              removeService={removeService}
              removeProduct={removeProduct}
              total={calculateTotal()}
            />
          </div>
        </div>
      </div>
    </div>
  );
}

function Step1ClientInfo({ formData, updateFormData, errors }: { formData: FormData; updateFormData: (field: keyof FormData, value: any) => void; errors: FormErrors }) {
  return (
    <div className="bg-card border border-border rounded-lg p-8">
      <h2 className="text-2xl mb-6 text-foreground">Данные клиента</h2>
      <div className="space-y-6">
        <div>
          <label htmlFor="client-name" className="block mb-2 text-foreground">
            ФИО *
          </label>
          <input
            id="client-name"
            type="text"
            required
            value={formData.clientName}
            onChange={(e) => updateFormData("clientName", e.target.value)}
            pattern="[\p{L}]+(?:[ '\-][\p{L}]+)*"
            className={`w-full px-4 py-3 bg-input-background border rounded-lg focus:outline-none focus:ring-2 focus:ring-ring ${
              errors.clientName ? "border-destructive" : "border-border"
            }`}
            placeholder="Иванов Иван Иванович"
          />
          {errors.clientName && (
            <p className="text-destructive text-sm mt-1">{errors.clientName}</p>
          )}
        </div>
        <div>
          <label htmlFor="client-phone" className="block mb-2 text-foreground">
            Номер телефона *
          </label>
          <input
            id="client-phone"
            type="tel"
            required
            value={formData.clientPhone}
            onChange={(e) => updateFormData("clientPhone", e.target.value)}
            className={`w-full px-4 py-3 bg-input-background border rounded-lg focus:outline-none focus:ring-2 focus:ring-ring ${
              errors.clientPhone ? "border-destructive" : "border-border"
            }`}
            placeholder="+7 (123) 456-78-90"
          />
          {errors.clientPhone && (
            <p className="text-destructive text-sm mt-1">{errors.clientPhone}</p>
          )}
        </div>
        <div>
          <label htmlFor="client-email" className="block mb-2 text-foreground">
            Email
          </label>
          <input
            id="client-email"
            type="email"
            value={formData.clientEmail}
            onChange={(e) => updateFormData("clientEmail", e.target.value)}
            className={`w-full px-4 py-3 bg-input-background border rounded-lg focus:outline-none focus:ring-2 focus:ring-ring ${
              errors.clientEmail ? "border-destructive" : "border-border"
            }`}
            placeholder="ivan@example.com"
          />
          {errors.clientEmail && (
            <p className="text-destructive text-sm mt-1">{errors.clientEmail}</p>
          )}
        </div>
      </div>
    </div>
  );
}

function Step2DeceasedInfo({ formData, updateFormData, errors }: { formData: FormData; updateFormData: (field: keyof FormData, value: any) => void; errors: FormErrors }) {
  return (
    <div className="bg-card border border-border rounded-lg p-8">
      <h2 className="text-2xl mb-6 text-foreground">Данные умершего</h2>
      <div className="space-y-6">
        <div>
          <label htmlFor="deceased-name" className="block mb-2 text-foreground">
            ФИО *
          </label>
          <input
            id="deceased-name"
            type="text"
            required
            value={formData.deceasedName}
            onChange={(e) => updateFormData("deceasedName", e.target.value)}
            pattern="[\p{L}]+(?:[ '\-][\p{L}]+)*"
            className={`w-full px-4 py-3 bg-input-background border rounded-lg focus:outline-none focus:ring-2 focus:ring-ring ${
              errors.deceasedName ? "border-destructive" : "border-border"
            }`}
            placeholder="Петрова Мария Ивановна"
          />
          {errors.deceasedName && (
            <p className="text-destructive text-sm mt-1">{errors.deceasedName}</p>
          )}
        </div>
        <div>
          <label htmlFor="date-of-birth" className="block mb-2 text-foreground">
            Дата рождения
          </label>
          <input
            id="date-of-birth"
            type="date"
            value={formData.dateOfBirth}
            onChange={(e) => updateFormData("dateOfBirth", e.target.value)}
            max={formData.dateOfDeath || undefined}
            className={`w-full px-4 py-3 bg-input-background border rounded-lg focus:outline-none focus:ring-2 focus:ring-ring ${
              errors.dateOfBirth ? "border-destructive" : "border-border"
            }`}
          />
          {errors.dateOfBirth && (
            <p className="text-destructive text-sm mt-1">{errors.dateOfBirth}</p>
          )}
        </div>
        <div>
          <label htmlFor="date-of-death" className="block mb-2 text-foreground">
            Дата смерти *
          </label>
          <input
            id="date-of-death"
            type="date"
            required
            value={formData.dateOfDeath}
            onChange={(e) => updateFormData("dateOfDeath", e.target.value)}
            min={formData.dateOfBirth || undefined}
            className={`w-full px-4 py-3 bg-input-background border rounded-lg focus:outline-none focus:ring-2 focus:ring-ring ${
              errors.dateOfDeath ? "border-destructive" : "border-border"
            }`}
          />
          {errors.dateOfDeath && (
            <p className="text-destructive text-sm mt-1">{errors.dateOfDeath}</p>
          )}
        </div>
      </div>
    </div>
  );
}

function Step3Services({ selectedServices, toggleService }: { selectedServices: string[]; toggleService: (id: string) => void }) {
  return (
    <div className="bg-card border border-border rounded-lg p-8">
      <h2 className="text-2xl mb-6 text-foreground">Выбор услуг</h2>
      <p className="text-muted-foreground mb-6">Выберите одну или несколько услуг *</p>
      {selectedServices.length === 0 && (
        <div className="mb-4 p-4 bg-destructive/10 border-2 border-destructive/50 rounded-lg">
          <div className="flex items-start gap-2">
            <svg className="w-5 h-5 text-destructive flex-shrink-0 mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
            </svg>
            <div>
              <p className="text-destructive font-medium">Пожалуйста, выберите хотя бы одну услугу</p>
              <p className="text-destructive text-xs mt-1">Для продолжения необходимо выбрать минимум одну услугу</p>
            </div>
          </div>
        </div>
      )}
      <div className="space-y-3">
        {availableServices.map((service) => {
          const isSelected = selectedServices.includes(service.id);
          return (
            <label
              key={service.id}
              className={`flex items-center gap-4 p-5 border-2 rounded-lg cursor-pointer transition-all ${
                isSelected
                  ? "border-primary bg-primary/5 shadow-sm"
                  : "border-border hover:bg-secondary"
              }`}
            >
              <input
                type="checkbox"
                checked={isSelected}
                onChange={() => toggleService(service.id)}
                className="w-5 h-5 accent-primary"
              />
              <CatalogItemImage
                kind="service"
                title={service.name}
                className="w-16 h-12 object-cover rounded border border-border bg-muted flex-shrink-0"
              />
              <span className="flex-1 text-foreground">{service.name}</span>
              <span className={isSelected ? "text-primary" : "text-muted-foreground"}>
                {service.price.toLocaleString()} ₽
              </span>
            </label>
          );
        })}
      </div>
    </div>
  );
}

function Step4Products({ selectedProducts, toggleProduct }: { selectedProducts: string[]; toggleProduct: (id: string) => void }) {
  return (
    <div className="bg-card border border-border rounded-lg p-8">
      <h2 className="text-2xl mb-6 text-foreground">Выбор товаров</h2>
      <p className="text-muted-foreground mb-6">Выберите необходимые товары (необязательно)</p>
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        {availableProducts.map((product) => {
          const isSelected = selectedProducts.includes(product.id);
          return (
            <div
              key={product.id}
              onClick={() => toggleProduct(product.id)}
              className={`p-5 border-2 rounded-lg cursor-pointer transition-all ${
                isSelected
                  ? "border-primary bg-primary/5 shadow-sm"
                  : "border-border hover:bg-secondary"
              }`}
            >
              <div className="flex items-start gap-3">
                <input
                  type="checkbox"
                  checked={isSelected}
                  onChange={() => {}}
                  className="w-5 h-5 accent-primary mt-0.5"
                />
                <CatalogItemImage
                  kind="product"
                  title={product.name}
                  className="w-20 h-16 object-cover rounded border border-border bg-muted flex-shrink-0"
                />
                <div className="flex-1">
                  <div className="text-foreground mb-2">{product.name}</div>
                  <div className={isSelected ? "text-primary" : "text-muted-foreground"}>
                    {product.price.toLocaleString()} ₽
                  </div>
                </div>
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
}

function Step5Details({
  formData,
  updateFormData,
  sections,
  plots,
  selectedPlotId,
  isLoadingPlots,
  plotsError,
  selectSection,
  selectPlot,
  paymentDetails,
  handleCardNumberChange,
  handleExpiryChange,
  updatePaymentDetails,
  errors,
}: {
  formData: FormData;
  updateFormData: (field: keyof FormData, value: any) => void;
  sections: CemeterySection[];
  plots: CemeteryPlot[];
  selectedPlotId: number | null;
  isLoadingPlots: boolean;
  plotsError: string | null;
  selectSection: (sectionName: string) => void;
  selectPlot: (plot: CemeteryPlot) => void;
  paymentDetails: PaymentDetails;
  handleCardNumberChange: (value: string) => void;
  handleExpiryChange: (value: string) => void;
  updatePaymentDetails: (field: keyof PaymentDetails, value: string) => void;
  errors: FormErrors;
}) {
  return (
    <div className="space-y-6">
      {/* Service Details */}
      <div className="bg-card border border-border rounded-lg p-8">
        <h2 className="text-2xl mb-6 text-foreground">Детали церемонии</h2>
        <div className="space-y-6">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div>
              <label htmlFor="service-date" className="block mb-2 text-foreground">
                Предпочтительная дата{formData.serviceType === "burial" ? " *" : ""}
              </label>
              <input
                id="service-date"
                type="date"
                value={formData.serviceDate}
                onChange={(e) => updateFormData("serviceDate", e.target.value)}
                className={`w-full px-4 py-3 bg-input-background border rounded-lg focus:outline-none focus:ring-2 focus:ring-ring ${
                  errors.serviceDate ? "border-destructive" : "border-border"
                }`}
              />
              {errors.serviceDate && (
                <p className="mt-2 text-sm text-destructive">{errors.serviceDate}</p>
              )}
              {formData.serviceType === "burial" && !errors.serviceDate && (
                <p className="mt-2 text-sm text-muted-foreground">
                  Дата обязательна для бронирования места.
                </p>
              )}
            </div>
            <div>
              <label htmlFor="service-time" className="block mb-2 text-foreground">
                Предпочтительное время
              </label>
              <input
                id="service-time"
                type="time"
                value={formData.serviceTime}
                onChange={(e) => updateFormData("serviceTime", e.target.value)}
                className="w-full px-4 py-3 bg-input-background border border-border rounded-lg focus:outline-none focus:ring-2 focus:ring-ring"
              />
            </div>
          </div>
          <div>
            <label htmlFor="service-address" className="block mb-2 text-foreground">
              Место проведения
            </label>
            <input
              id="service-address"
              type="text"
              value={formData.serviceAddress}
              onChange={(e) => updateFormData("serviceAddress", e.target.value)}
              className="w-full px-4 py-3 bg-input-background border border-border rounded-lg focus:outline-none focus:ring-2 focus:ring-ring"
              placeholder="Адрес или название учреждения"
            />
          </div>
        </div>
      </div>

      {/* Cemetery Information */}
      <div className="bg-card border border-border rounded-lg p-8">
        <h2 className="text-2xl mb-6 text-foreground">Информация о кладбище</h2>
        <div className="space-y-6">
          <div>
            <label htmlFor="service-type" className="block mb-2 text-foreground">
              Тип услуги
            </label>
            <select
              id="service-type"
              value={formData.serviceType}
              onChange={(e) => updateFormData("serviceType", e.target.value)}
              className="w-full px-4 py-3 bg-input-background border border-border rounded-lg focus:outline-none focus:ring-2 focus:ring-ring"
            >
              <option value="">Выберите тип услуги...</option>
              <option value="burial">Погребение</option>
              <option value="cremation">Кремация</option>
            </select>
          </div>

          {formData.serviceType === "burial" && (
            <div className="space-y-5">
              <div>
                <label htmlFor="cemetery-section" className="block mb-2 text-foreground">
                  Секция кладбища
                </label>
                <select
                  id="cemetery-section"
                  value={formData.selectedSectionName}
                  onChange={(event) => selectSection(event.target.value)}
                  className="w-full px-4 py-3 bg-input-background border border-border rounded-lg focus:outline-none focus:ring-2 focus:ring-ring"
                >
                  <option value="">Выберите секцию...</option>
                  {sections.map((section) => (
                    <option key={section.id} value={section.name}>
                      {section.name}
                    </option>
                  ))}
                </select>
              </div>

              <div className="flex items-center justify-between mb-3">
                <label className="block text-foreground">
                  Место захоронения
                </label>
                {formData.selectedPlotLabel && (
                  <span className="text-sm text-primary">
                    Выбрано: {formData.selectedPlotLabel}
                  </span>
                )}
              </div>

              {!formData.selectedSectionName && !plotsError && (
                <div className="bg-secondary/50 rounded-lg p-5 text-muted-foreground">
                  Сначала выберите секцию кладбища.
                </div>
              )}

              {isLoadingPlots && (
                <div className="bg-secondary/50 rounded-lg p-5 text-muted-foreground">
                  Загрузка доступных мест...
                </div>
              )}

              {plotsError && (
                <div className="bg-destructive/10 border border-destructive/40 rounded-lg p-4 text-destructive">
                  {plotsError}
                </div>
              )}

              {formData.selectedSectionName && !isLoadingPlots && !plotsError && (
                <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
                  {plots.map((plot) => {
                    const isSelected = selectedPlotId === plot.id;

                    return (
                      <button
                        key={plot.id}
                        type="button"
                        disabled={!plot.available}
                        onClick={() => selectPlot(plot)}
                        className={`text-left border rounded-lg p-5 transition-all ${
                          isSelected
                            ? "border-primary bg-primary/10"
                            : "border-border bg-secondary/30 hover:bg-secondary"
                        } ${!plot.available ? "opacity-45 cursor-not-allowed hover:bg-secondary/30" : ""}`}
                      >
                        <div className="text-xl text-foreground mb-2">{plot.label}</div>
                        <div className={`text-sm ${plot.available ? "text-primary" : "text-muted-foreground"}`}>
                          {plot.available ? "Доступно" : "Недоступно"}
                        </div>
                      </button>
                    );
                  })}
                </div>
              )}

              {errors.cemeteryPlot && (
                <p className="mt-3 text-sm text-destructive">{errors.cemeteryPlot}</p>
              )}
              {selectedPlotId === null && !errors.cemeteryPlot && (
                <p className="mt-3 text-sm text-muted-foreground">
                  Выберите свободное место, чтобы продолжить оформление.
                </p>
              )}
            </div>
          )}

          {formData.serviceType === "cremation" && (
            <div className="bg-secondary/50 border border-border rounded-lg p-4 text-sm text-muted-foreground">
              Для кремации место захоронения не резервируется. При необходимости укажите пожелания по урне, хранению или передаче праха в примечании.
            </div>
          )}

          <div>
            <label htmlFor="cemetery-notes" className="block mb-2 text-foreground">
              {formData.serviceType === "cremation" ? "Примечание по кремации" : "Примечания"}
            </label>
            <textarea
              id="cemetery-notes"
              rows={3}
              value={formData.cemeteryNotes}
              onChange={(e) => updateFormData("cemeteryNotes", e.target.value)}
              className="w-full px-4 py-3 bg-input-background border border-border rounded-lg focus:outline-none focus:ring-2 focus:ring-ring resize-none"
              placeholder={formData.serviceType === "cremation"
                ? "Например: пожелания по урне или передаче праха"
                : "Особые пожелания по месту захоронения"}
            />
          </div>
        </div>
      </div>

      {/* Payment Method */}
      <div className="bg-card border border-border rounded-lg p-8">
        <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3 mb-6">
          <h2 className="text-2xl text-foreground">Способ оплаты</h2>
          <div className="flex items-center gap-2">
            <span className="text-sm text-muted-foreground">Статус оплаты:</span>
            <span className="text-sm px-3 py-1 bg-secondary text-secondary-foreground rounded-full border border-border">
              Ожидает оплаты
            </span>
          </div>
        </div>
        <p className="text-sm text-muted-foreground mb-6">
          Выберите оплату сейчас или сохраните заказ и оплатите его позже в личном кабинете.
        </p>
        <div className="space-y-4">
          <label className="flex items-center gap-3 p-4 border border-border rounded-lg cursor-pointer hover:bg-secondary transition-colors">
            <input
              type="radio"
              name="payment"
              value="credit"
              checked={formData.paymentMethod === "credit"}
              onChange={(e) => updateFormData("paymentMethod", e.target.value)}
              className="w-5 h-5 accent-primary"
            />
            <div className="flex-1">
              <div className="text-foreground">Банковская карта</div>
              <div className="text-sm text-muted-foreground">
                Оплатить заказ сразу после его создания
              </div>
            </div>
          </label>

          {formData.paymentMethod === "credit" && (
            <div className="border border-border rounded-lg p-5 space-y-4">
              <div>
                <label htmlFor="order-card-number" className="block text-sm text-foreground mb-2">
                  Номер карты
                </label>
                <input
                  id="order-card-number"
                  value={paymentDetails.cardNumber}
                  onChange={(event) => handleCardNumberChange(event.target.value)}
                  inputMode="numeric"
                  autoComplete="cc-number"
                  placeholder="0000 0000 0000 0000"
                  required
                  className="w-full h-11 px-3 bg-input-background border border-border rounded-md text-foreground outline-none focus:ring-2 focus:ring-ring"
                />
              </div>

              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                <div>
                  <label htmlFor="order-expiry-date" className="block text-sm text-foreground mb-2">
                    Срок действия
                  </label>
                  <input
                    id="order-expiry-date"
                    value={paymentDetails.expiryDate}
                    onChange={(event) => handleExpiryChange(event.target.value)}
                    inputMode="numeric"
                    autoComplete="cc-exp"
                    placeholder="ММ/ГГ"
                    required
                    className="w-full h-11 px-3 bg-input-background border border-border rounded-md text-foreground outline-none focus:ring-2 focus:ring-ring"
                  />
                </div>
                <div>
                  <label htmlFor="order-card-cvv" className="block text-sm text-foreground mb-2">
                    CVV
                  </label>
                  <input
                    id="order-card-cvv"
                    type="password"
                    value={paymentDetails.cvv}
                    onChange={(event) =>
                      updatePaymentDetails(
                        "cvv",
                        event.target.value.replace(/\D/g, "").slice(0, 3),
                      )
                    }
                    inputMode="numeric"
                    autoComplete="cc-csc"
                    placeholder="000"
                    required
                    className="w-full h-11 px-3 bg-input-background border border-border rounded-md text-foreground outline-none focus:ring-2 focus:ring-ring"
                  />
                </div>
              </div>
              <p className="text-xs text-muted-foreground">
                Данные карты используются только для выполнения платежа и не сохраняются в заказе.
              </p>
            </div>
          )}

          <label className="flex items-center gap-3 p-4 border border-border rounded-lg cursor-pointer hover:bg-secondary transition-colors">
            <input
              type="radio"
              name="payment"
              value="later"
              checked={formData.paymentMethod === "later"}
              onChange={(e) => updateFormData("paymentMethod", e.target.value)}
              className="w-5 h-5 accent-primary"
            />
            <div className="flex-1">
              <div className="text-foreground">Оплатить позже</div>
              <div className="text-sm text-muted-foreground">
                Заказ появится в личном кабинете, где его можно будет оплатить картой
              </div>
            </div>
          </label>
        </div>
      </div>

      {/* Additional Comments */}
      <div className="bg-card border border-border rounded-lg p-8">
        <h2 className="text-2xl mb-6 text-foreground">Дополнительная информация</h2>
        <label htmlFor="comments" className="block mb-2 text-foreground">
          Особые пожелания или комментарии
        </label>
        <textarea
          id="comments"
          rows={5}
          value={formData.comments}
          onChange={(e) => updateFormData("comments", e.target.value)}
          className="w-full px-4 py-3 bg-input-background border border-border rounded-lg focus:outline-none focus:ring-2 focus:ring-ring resize-none"
          placeholder="Укажите особые пожелания или дополнительную информацию..."
        />
      </div>
    </div>
  );
}

function Step6Review({ formData, jumpToStep }: { formData: FormData; jumpToStep: (step: number) => void }) {
  const selectedServicesList = availableServices.filter(s =>
    formData.selectedServices.includes(s.id)
  );
  const selectedProductsList = availableProducts.filter(p =>
    formData.selectedProducts.includes(p.id)
  );

  const getPaymentHelperText = () => {
    switch (formData.paymentMethod) {
      case "credit":
        return "Оплата картой будет выполнена сразу после создания заказа";
      case "later":
        return "Заказ можно будет оплатить позже в личном кабинете";
      default:
        return "Оплата ожидается";
    }
  };

  return (
    <div className="bg-card border border-border rounded-lg p-8">
      <h2 className="text-2xl mb-6 text-foreground">Проверка заказа</h2>

      {formData.selectedServices.length === 0 && (
        <div className="mb-6 p-4 bg-destructive/10 border-2 border-destructive/50 rounded-lg">
          <div className="flex items-start gap-2">
            <svg className="w-5 h-5 text-destructive flex-shrink-0 mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
            </svg>
            <div>
              <p className="text-destructive font-medium">Невозможно отправить заказ</p>
              <p className="text-destructive text-sm mt-1">Пожалуйста, выберите хотя бы одну услугу перед отправкой заказа</p>
            </div>
          </div>
        </div>
      )}

      <div className="space-y-6">
        {/* Client Information */}
        <div>
          <div className="flex items-center justify-between mb-3">
            <h3 className="text-foreground">Данные клиента</h3>
            <button
              onClick={() => jumpToStep(1)}
              className="text-sm text-primary hover:underline"
            >
              Изменить
            </button>
          </div>
          <div className="bg-secondary/50 rounded-lg p-4 space-y-1 text-sm">
            <p className="text-foreground"><span className="text-muted-foreground">ФИО:</span> {formData.clientName}</p>
            <p className="text-foreground"><span className="text-muted-foreground">Телефон:</span> {formData.clientPhone}</p>
            {formData.clientEmail && (
              <p className="text-foreground"><span className="text-muted-foreground">Email:</span> {formData.clientEmail}</p>
            )}
          </div>
        </div>

        {/* Deceased Information */}
        <div>
          <div className="flex items-center justify-between mb-3">
            <h3 className="text-foreground">Данные умершего</h3>
            <button
              onClick={() => jumpToStep(2)}
              className="text-sm text-primary hover:underline"
            >
              Изменить
            </button>
          </div>
          <div className="bg-secondary/50 rounded-lg p-4 space-y-1 text-sm">
            <p className="text-foreground"><span className="text-muted-foreground">ФИО:</span> {formData.deceasedName}</p>
            {formData.dateOfBirth && (
              <p className="text-foreground"><span className="text-muted-foreground">Дата рождения:</span> {formData.dateOfBirth}</p>
            )}
            <p className="text-foreground"><span className="text-muted-foreground">Дата смерти:</span> {formData.dateOfDeath}</p>
          </div>
        </div>

        {/* Selected Services */}
        <div>
          <div className="flex items-center justify-between mb-3">
            <h3 className="text-foreground">Выбранные услуги</h3>
            <button
              onClick={() => jumpToStep(3)}
              className="text-sm text-primary hover:underline"
            >
              Изменить
            </button>
          </div>
          {selectedServicesList.length > 0 ? (
            <div className="bg-secondary/50 rounded-lg p-4 space-y-2">
              {selectedServicesList.map(service => (
                <div key={service.id} className="flex justify-between text-sm">
                  <span className="text-foreground">{service.name}</span>
                  <span className="text-primary">{service.price.toLocaleString()} ₽</span>
                </div>
              ))}
            </div>
          ) : (
            <div className="bg-destructive/10 border-2 border-destructive/50 rounded-lg p-4">
              <p className="text-destructive font-medium text-sm">Услуги не выбраны</p>
              <p className="text-destructive text-xs mt-1">Необходимо выбрать хотя бы одну услугу</p>
            </div>
          )}
        </div>

        {/* Selected Products */}
        <div>
          <div className="flex items-center justify-between mb-3">
            <h3 className="text-foreground">Выбранные товары</h3>
            <button
              onClick={() => jumpToStep(4)}
              className="text-sm text-primary hover:underline"
            >
              Изменить
            </button>
          </div>
          {selectedProductsList.length > 0 ? (
            <div className="bg-secondary/50 rounded-lg p-4 space-y-2">
              {selectedProductsList.map(product => (
                <div key={product.id} className="flex justify-between text-sm">
                  <span className="text-foreground">{product.name}</span>
                  <span className="text-primary">{product.price.toLocaleString()} ₽</span>
                </div>
              ))}
            </div>
          ) : (
            <div className="bg-secondary/50 rounded-lg p-4 text-sm text-muted-foreground">
              Товары не выбраны
            </div>
          )}
        </div>

        {/* Service Details */}
        <div>
          <div className="flex items-center justify-between mb-3">
            <h3 className="text-foreground">Детали церемонии</h3>
            <button
              onClick={() => jumpToStep(5)}
              className="text-sm text-primary hover:underline"
            >
              Изменить
            </button>
          </div>
          <div className="bg-secondary/50 rounded-lg p-4 space-y-1 text-sm">
            {formData.serviceType && (
              <p className="text-foreground">
                <span className="text-muted-foreground">Тип услуги:</span>{" "}
                {formData.serviceType === "burial" ? "Погребение" : formData.serviceType === "cremation" ? "Кремация" : formData.serviceType}
              </p>
            )}
            {formData.serviceDate && (
              <p className="text-foreground">
                <span className="text-muted-foreground">Дата:</span> {formData.serviceDate}
                {formData.serviceTime && ` в ${formData.serviceTime}`}
              </p>
            )}
            {formData.serviceAddress && (
              <p className="text-foreground"><span className="text-muted-foreground">Место:</span> {formData.serviceAddress}</p>
            )}
            {formData.cemetery && (
              <p className="text-foreground">
                <span className="text-muted-foreground">Кладбище:</span>{" "}
                {formData.cemetery === "Другое (указать)" ? formData.cemeteryOther : formData.cemetery}
              </p>
            )}
            {formData.cemeteryNotes && (
              <p className="text-muted-foreground">{formData.cemeteryNotes}</p>
            )}
            {!formData.serviceType && !formData.serviceDate && !formData.serviceAddress && !formData.cemetery && (
              <p className="text-muted-foreground">Дополнительные детали не указаны</p>
            )}
          </div>
        </div>

        {formData.serviceType === "burial" && (
        <div>
          <div className="flex items-center justify-between mb-3">
            <h3 className="text-foreground">Место захоронения</h3>
            <button
              onClick={() => jumpToStep(5)}
              className="text-sm text-primary hover:underline"
            >
              Изменить
            </button>
          </div>
          <div className="bg-secondary/50 rounded-lg p-4 text-sm">
            {formData.selectedPlotLabel ? (
              <p className="text-foreground">{formData.selectedPlotLabel}</p>
            ) : (
              <p className="text-muted-foreground">Место не выбрано</p>
            )}
          </div>
        </div>
        )}

        {/* Payment Method */}
        <div>
          <h3 className="mb-3 text-foreground">Способ оплаты</h3>
          <div className="bg-secondary/50 rounded-lg p-4 text-sm space-y-1">
            <p className="text-foreground">
              {formData.paymentMethod === "credit" && "Банковская карта"}
              {formData.paymentMethod === "later" && "Оплатить позже"}
            </p>
            <p className="text-muted-foreground text-xs">Статус: Ожидает оплаты</p>
            <p className="text-muted-foreground text-xs">{getPaymentHelperText()}</p>
          </div>
        </div>

        {/* Additional Comments */}
        {formData.comments && (
          <div>
            <h3 className="mb-3 text-foreground">Дополнительные комментарии</h3>
            <div className="bg-secondary/50 rounded-lg p-4 text-sm">
              <p className="text-foreground">{formData.comments}</p>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}

function OrderSummary({ selectedServices, selectedProducts, removeService, removeProduct, total }: { selectedServices: string[]; selectedProducts: string[]; removeService: (id: string) => void; removeProduct: (id: string) => void; total: number }) {
  const selectedServicesList = availableServices.filter(s => selectedServices.includes(s.id));
  const selectedProductsList = availableProducts.filter(p => selectedProducts.includes(p.id));
  const isLastService = selectedServicesList.length === 1;

  const handleRemoveService = (id: string) => {
    if (isLastService) {
      return; // Не удаляем последнюю услугу
    }
    removeService(id);
  };

  return (
    <div className="bg-card border border-border rounded-lg p-6 sticky top-6">
      <h3 className="text-xl mb-4 text-foreground">Сводка заказа</h3>

      <div className="space-y-4 mb-6">
        {selectedServicesList.length > 0 ? (
          <div>
            <div className="text-sm text-muted-foreground mb-2">Услуги</div>
            <div className="space-y-2">
              {selectedServicesList.map(service => (
                <div key={service.id} className="flex items-start justify-between gap-2 text-sm">
                  <span className="text-foreground flex-1">{service.name}</span>
                  <span className="text-primary whitespace-nowrap">{service.price.toLocaleString()} ₽</span>
                  {!isLastService ? (
                    <button
                      onClick={() => handleRemoveService(service.id)}
                      className="text-muted-foreground hover:text-destructive transition-colors"
                      title="Удалить"
                    >
                      <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                      </svg>
                    </button>
                  ) : (
                    <button
                      disabled
                      className="text-muted-foreground/30 cursor-not-allowed"
                      title="Требуется минимум одна услуга"
                    >
                      <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                      </svg>
                    </button>
                  )}
                </div>
              ))}
            </div>
          </div>
        ) : (
          <div>
            <div className="text-sm text-muted-foreground mb-2">Услуги</div>
            <div className="p-4 bg-destructive/10 border-2 border-destructive/50 rounded-lg">
              <p className="text-destructive text-sm font-medium">Услуги не выбраны</p>
              <p className="text-destructive text-xs mt-1">Выберите хотя бы одну услугу</p>
            </div>
          </div>
        )}

        {selectedProductsList.length > 0 ? (
          <div>
            <div className="text-sm text-muted-foreground mb-2">Товары</div>
            <div className="space-y-2">
              {selectedProductsList.map(product => (
                <div key={product.id} className="flex items-start justify-between gap-2 text-sm">
                  <span className="text-foreground flex-1">{product.name}</span>
                  <span className="text-primary whitespace-nowrap">{product.price.toLocaleString()} ₽</span>
                  <button
                    onClick={() => removeProduct(product.id)}
                    className="text-muted-foreground hover:text-destructive transition-colors"
                    title="Удалить"
                  >
                    <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                    </svg>
                  </button>
                </div>
              ))}
            </div>
          </div>
        ) : (
          <div>
            <div className="text-sm text-muted-foreground mb-2">Товары</div>
            <div className="text-sm text-muted-foreground text-center py-3 bg-secondary/50 rounded-lg">
              Товары не выбраны
            </div>
          </div>
        )}
      </div>

      <div className="pt-4 border-t border-border">
        <div className="flex justify-between items-center">
          <span className="text-lg text-foreground">Итого</span>
          <span className="text-2xl text-primary">{total.toLocaleString()} ₽</span>
        </div>
        <p className="text-xs text-muted-foreground mt-2">
          * Предварительная стоимость, окончательная сумма будет подтверждена
        </p>
      </div>
    </div>
  );
}

function OrderConfirmation({
  orderId,
  paymentResult,
  paymentError,
  onNewOrder,
}: {
  orderId: string;
  paymentResult: "paid" | "deferred" | "failed" | null;
  paymentError: string;
  onNewOrder: () => void;
}) {
  const navigate = useNavigate();
  const [copied, setCopied] = useState(false);

  const copyOrderId = () => {
    navigator.clipboard.writeText(orderId).then(() => {
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    });
  };

  const handleGoHome = () => {
    onNewOrder();
    navigate("/");
  };

  return (
    <div className="min-h-screen bg-background flex items-center justify-center py-12 px-6">
      <div className="max-w-2xl w-full">
        <div className="bg-card border border-border rounded-lg p-12 text-center">
          <div className="w-20 h-20 bg-primary/10 rounded-full flex items-center justify-center mx-auto mb-6">
            <svg
              className="w-10 h-10 text-primary"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M5 13l4 4L19 7"
              />
            </svg>
          </div>

          <h1 className="text-3xl mb-4 text-foreground">Заказ успешно отправлен</h1>
          <p className="text-lg text-muted-foreground mb-8">
            Благодарим за обращение. Мы свяжемся с вами в ближайшее время для подтверждения всех деталей.
          </p>

          <div className="bg-secondary/50 border border-border rounded-lg p-6 mb-2">
            <div className="text-sm text-muted-foreground mb-2">Номер заказа</div>
            <div className="flex items-center justify-center gap-3">
              <div className="text-3xl text-primary tracking-wider font-medium">{orderId}</div>
              <button
                onClick={copyOrderId}
                className="text-muted-foreground hover:text-primary transition-colors p-2 hover:bg-secondary rounded"
                title="Скопировать номер заказа"
              >
                {copied ? (
                  <svg className="w-5 h-5 text-primary" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                  </svg>
                ) : (
                  <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
                  </svg>
                )}
              </button>
            </div>
          </div>

          <div className="flex justify-center mb-8">
            <span className={`inline-flex items-center px-3 py-1 rounded-full text-sm border ${
              paymentResult === "paid"
                ? "bg-primary/10 text-primary border-primary/30"
                : "bg-secondary text-secondary-foreground border-border"
            }`}>
              <span className={`w-2 h-2 rounded-full mr-2 ${
                paymentResult === "paid" ? "bg-primary" : "bg-muted-foreground"
              }`}></span>
              {paymentResult === "paid" ? "Статус: Оплачен" : "Статус: Ожидает оплаты"}
            </span>
          </div>

          {paymentResult === "paid" && (
            <div className="mb-8 border border-primary/30 bg-primary/10 rounded-lg p-4 text-left">
              <p className="text-primary">Оплата банковской картой прошла успешно.</p>
            </div>
          )}

          {paymentResult === "failed" && (
            <div className="mb-8 border border-destructive/30 bg-destructive/10 rounded-lg p-4 text-left">
              <p className="text-destructive">Заказ создан, но оплата не прошла.</p>
              <p className="text-sm text-muted-foreground mt-1">
                {paymentError || "Повторить оплату можно в личном кабинете."}
              </p>
            </div>
          )}

          {paymentResult === "deferred" && (
            <div className="mb-8 border border-border bg-secondary/30 rounded-lg p-4 text-left">
              <p className="text-foreground">Вы выбрали оплату позже.</p>
              <p className="text-sm text-muted-foreground mt-1">
                Оплатить заказ можно в личном кабинете по номеру телефона из заказа.
              </p>
            </div>
          )}

          <div className="space-y-4 text-left mb-8 bg-secondary/30 rounded-lg p-6">
            <h2 className="text-xl text-foreground">Что будет дальше:</h2>
            <ul className="space-y-3 text-muted-foreground">
              <li className="flex gap-3">
                <span className="text-primary flex-shrink-0 w-6 h-6 rounded-full bg-primary/10 flex items-center justify-center text-sm">1</span>
                <span>Наш специалист свяжется с вами в течение 24 часов для уточнения деталей</span>
              </li>
              <li className="flex gap-3">
                <span className="text-primary flex-shrink-0 w-6 h-6 rounded-full bg-primary/10 flex items-center justify-center text-sm">2</span>
                <span>Мы обсудим все нюансы организации и согласуем окончательную стоимость</span>
              </li>
              <li className="flex gap-3">
                <span className="text-primary flex-shrink-0 w-6 h-6 rounded-full bg-primary/10 flex items-center justify-center text-sm">3</span>
                <span>После согласования вы получите договор на оказание услуг</span>
              </li>
            </ul>
          </div>

          <div className="flex flex-col sm:flex-row gap-4 mb-6">
            <button
              onClick={handleGoHome}
              className="flex-1 bg-secondary text-secondary-foreground border border-border py-3 px-6 rounded-lg hover:bg-accent transition-colors"
            >
              На главную
            </button>
            <a
              href="tel:+79991234567"
              className="flex-1 bg-primary text-primary-foreground py-3 px-6 rounded-lg hover:opacity-90 transition-opacity flex items-center justify-center gap-2"
            >
              <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 5a2 2 0 012-2h3.28a1 1 0 01.948.684l1.498 4.493a1 1 0 01-.502 1.21l-2.257 1.13a11.042 11.042 0 005.516 5.516l1.13-2.257a1 1 0 011.21-.502l4.493 1.498a1 1 0 01.684.949V19a2 2 0 01-2 2h-1C9.716 21 3 14.284 3 6V5z" />
              </svg>
              <span>+7 (999) 123-45-67</span>
            </a>
          </div>

          <div className="border-t border-border pt-6">
            <p className="text-sm text-muted-foreground">
              <svg className="w-4 h-4 inline mr-1 -mt-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
              Если у вас срочный вопрос — свяжитесь с нами по телефону круглосуточно
            </p>
          </div>
        </div>
      </div>
    </div>
  );
}
