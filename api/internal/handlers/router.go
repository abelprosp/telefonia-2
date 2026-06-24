package handlers



import (

	"encoding/json"

	"net/http"



	"github.com/go-chi/chi/v5"

	"github.com/luxus-connect/telefonia/api/internal/httputil"

	"github.com/luxus-connect/telefonia/api/internal/models"

	"github.com/luxus-connect/telefonia/api/internal/notifications"

	"github.com/luxus-connect/telefonia/api/internal/services"

)



type Handler struct {

	Svc       *services.Service

	Presigned *services.PresignedService

}



func decodeJSON(r *http.Request, v any) error {

	defer r.Body.Close()

	dec := json.NewDecoder(r.Body)

	dec.DisallowUnknownFields()

	return dec.Decode(v)

}



func (h *Handler) RegisterRoutes(

	r chi.Router,

	auth func(http.Handler) http.Handler,

	operational func(http.Handler) http.Handler,

	financial func(http.Handler) http.Handler,

	master func(http.Handler) http.Handler,

	partner func(http.Handler) http.Handler,

) {

	r.Route("/v1", func(r chi.Router) {

		r.Post("/webhooks/sicredi", h.sicrediWebhook)

		r.Group(func(r chi.Router) {

			r.Use(auth)

			r.Post("/pre-signed-urls/upload", h.postPresignedUpload)

			r.Post("/pre-signed-urls/download", h.postPresignedDownload)

		})



		r.Group(func(r chi.Router) {

			r.Use(auth, operational)



			r.Get("/stats/dashboard", h.getDashboardStats)



			r.Route("/providers", func(r chi.Router) {

				r.Get("/", h.listProviders)

				r.Post("/", h.createProvider)

				r.Route("/{id}", func(r chi.Router) {

					r.Get("/", h.getProvider)

					r.Patch("/", h.updateProvider)

					r.Delete("/", h.inactivateProvider)

					r.Post("/plans", h.createProviderPlan)

					r.Patch("/plans/{planId}", h.updateProviderPlan)

				})

			})



			r.Route("/provider-invoices", func(r chi.Router) {

				r.Get("/", h.listProviderInvoices)

				r.Post("/", h.requestProviderInvoiceImport)

				r.Get("/{id}", h.getProviderInvoice)

			})



			r.Route("/customers", func(r chi.Router) {

				r.Get("/", h.listCustomers)

				r.Post("/", h.createCustomer)

				r.Route("/{id}", func(r chi.Router) {

					r.Get("/", h.getCustomer)

					r.Patch("/", h.updateCustomer)

					r.Delete("/", h.inactivateCustomer)

					r.Get("/phone-lines", h.listCustomerPhoneLines)

					r.Get("/devices", h.listCustomerDevices)

					r.Post("/devices", h.assignCustomerDevice)

					r.Patch("/devices/{deviceLinkId}", h.updateCustomerDevice)

					r.Delete("/devices/{deviceLinkId}", h.unassignCustomerDevice)

					r.Get("/provider-links", h.listCustomerProviderLinks)

					r.Get("/attachments", h.listCustomerAttachments)

					r.Post("/attachments", h.createCustomerAttachment)

					r.Delete("/attachments/{attachmentId}", h.deleteCustomerAttachment)

					r.Get("/processing-months/{processingMonthId}/billing-readiness", h.getBillingReadiness)

					r.Post("/processing-months/{processingMonthId}/manual-release", h.manualReleaseCustomer)

					r.Post("/generate-billing-document", h.generateCustomerBillingDocument)

				})

			})



			r.Route("/phone-lines", func(r chi.Router) {

				r.Get("/", h.listPhoneLines)

				r.Post("/stock", h.createStockPhoneLine)

				r.Route("/{id}", func(r chi.Router) {

					r.Get("/", h.getPhoneLine)

					r.Get("/customer-links", h.listPhoneLineCustomerLinks)

					r.Post("/customer-links", h.assignPhoneLineCustomer)

					r.Post("/customer-links/transfer", h.transferPhoneLineCustomer)

					r.Patch("/customer-links/active", h.updateActivePhoneLineCustomerLink)

					r.Delete("/customer-links/active", h.unassignPhoneLineCustomer)

				})

			})



			r.Route("/device-stock", func(r chi.Router) {

				r.Get("/", h.listDeviceStockItems)

				r.Post("/", h.createDeviceStockItem)

				r.Route("/{id}", func(r chi.Router) {

					r.Get("/", h.getDeviceStockItem)

					r.Patch("/", h.updateDeviceStockItem)

				})

			})



			r.Route("/billing-cycles", func(r chi.Router) {

				r.Get("/", h.listBillingCycles)

				r.Post("/", h.createBillingCycle)

				r.Route("/{id}", func(r chi.Router) {

					r.Get("/", h.getBillingCycle)

					r.Patch("/", h.updateBillingCycle)

				})

			})



			r.Route("/processing-months", func(r chi.Router) {

				r.Get("/", h.listProcessingMonths)

				r.Post("/", h.createProcessingMonth)

				r.Route("/{id}", func(r chi.Router) {

					r.Get("/", h.getProcessingMonth)

					r.Post("/close", h.closeProcessingMonth)

					r.Post("/close-contingency", h.closeProcessingMonthContingency)

				})

			})



			r.Get("/cost-centers", h.listCostCenters)



			r.Route("/phone-line-operation-requests", func(r chi.Router) {

				r.Get("/", h.listLineOperationRequests)

				r.Patch("/{id}", h.reviewLineOperationRequest)

			})



			r.Route("/contract-templates", func(r chi.Router) {

				r.Get("/", h.listContractTemplates)

				r.Post("/", h.createContractTemplate)

				r.Route("/{id}", func(r chi.Router) {

					r.Get("/", h.getContractTemplate)

					r.Patch("/", h.updateContractTemplate)

				})

			})



			r.Route("/sales", func(r chi.Router) {

				r.Get("/", h.listSales)

				r.Post("/", h.createSale)

				r.Route("/{id}", func(r chi.Router) {

					r.Get("/", h.getSale)

					r.Patch("/", h.updateSale)

					r.Post("/confirm", h.confirmSale)

					r.Post("/cancel", h.cancelSale)

					r.Post("/items", h.addSaleLineItem)

					r.Delete("/items/{itemId}", h.deleteSaleLineItem)

				})

			})

		})



		r.Group(func(r chi.Router) {

			r.Use(auth, financial)



			r.Get("/financial/summary", h.getFinancialSummary)



			r.Route("/accounts-payable", func(r chi.Router) {

				r.Get("/", h.listAccountsPayable)

				r.Post("/", h.createAccountPayable)

				r.Post("/from-provider-invoice/{invoiceId}", h.createAccountPayableFromInvoice)

				r.Route("/{id}", func(r chi.Router) {

					r.Patch("/", h.updateAccountPayable)

					r.Post("/payments", h.registerPayablePayment)

				})

			})



			r.Route("/accounts-receivable", func(r chi.Router) {

				r.Get("/", h.listAccountsReceivable)

				r.Post("/", h.createAccountReceivable)

				r.Route("/{id}", func(r chi.Router) {

					r.Patch("/", h.updateAccountReceivable)

					r.Post("/payments", h.registerReceivablePayment)

				})

			})



			r.Route("/partner-sales", func(r chi.Router) {

				r.Get("/", h.listPartnerSales)

				r.Post("/sync", h.syncPartnerSales)

				r.Patch("/{id}", h.updatePartnerSaleStatus)

			})



			r.Route("/partner-commission-settings", func(r chi.Router) {

				r.Get("/", h.getPartnerCommissionSettings)

				r.Put("/", h.updatePartnerCommissionSettings)

			})



			r.Route("/invoice-email-templates", func(r chi.Router) {

				r.Get("/", h.listInvoiceEmailTemplates)

				r.Post("/", h.createInvoiceEmailTemplate)

				r.Route("/{id}", func(r chi.Router) {

					r.Get("/", h.getInvoiceEmailTemplate)

					r.Patch("/", h.updateInvoiceEmailTemplate)

				})

			})



			r.Route("/invoice-layout-templates", func(r chi.Router) {

				r.Get("/", h.listInvoiceLayoutTemplates)

				r.Post("/", h.createInvoiceLayoutTemplate)

				r.Post("/preview", h.previewInvoiceLayout)

				r.Route("/{id}", func(r chi.Router) {

					r.Get("/", h.getInvoiceLayoutTemplate)

					r.Patch("/", h.updateInvoiceLayoutTemplate)

				})

			})



			r.Route("/customer-billing-documents", func(r chi.Router) {

				r.Get("/", h.listCustomerBillingDocuments)

				r.Get("/bulk-preview", h.bulkBillingPreview)

				r.Get("/manual-preview", h.manualBillingPreview)

				r.Post("/bulk-generate", h.bulkGenerateBillingDocuments)

				r.Post("/manual-generate", h.manualGenerateBillingDocuments)

				r.Post("/from-receivable/{receivableId}", h.createCustomerBillingDocumentFromReceivable)

				r.Route("/{id}", func(r chi.Router) {

					r.Get("/", h.getCustomerBillingDocument)

					r.Get("/download", h.downloadCustomerBillingDocument)

					r.Patch("/", h.updateCustomerBillingDocument)

					r.Post("/issue-boleto", h.issueSicrediBoleto)

					r.Get("/boleto-pdf", h.getSicrediBoletoPDF)

					r.Post("/cancel-boleto", h.cancelSicrediBoleto)

					r.Patch("/boleto-due-date", h.alterSicrediBoletoDueDate)

					r.Post("/sync-payment", h.syncSicrediPayment)

					r.Post("/send", h.sendCustomerBillingDocument)

					r.Get("/send-log", h.listCustomerBillingSendLog)

				})

			})



			r.Route("/collections", func(r chi.Router) {

				r.Get("/overdue", h.listOverdueReceivables)

				r.Post("/remind", h.sendCollectionReminder)

				r.Post("/sync-sicredi-payments", h.syncSicrediPayments)

				r.Get("/sicredi/status", h.getSicrediStatus)

				r.Post("/sicredi/test-connection", h.testSicrediConnection)

				r.Post("/sicredi/register-webhook", h.registerSicrediWebhook)

			})

		})



		r.Group(func(r chi.Router) {

			r.Use(auth, master)



			r.Route("/users", func(r chi.Router) {

				r.Get("/", h.listOrganizationUsers)

				r.Post("/", h.createOrganizationUser)

				r.Patch("/{id}", h.updateOrganizationUser)

			})

		})



		r.Group(func(r chi.Router) {

			r.Use(auth, partner)



			r.Get("/partner/stats/dashboard", h.partnerGetDashboardStats)



			r.Route("/partner/providers", func(r chi.Router) {

				r.Get("/", h.partnerListProviders)

			})



			r.Route("/partner/customers", func(r chi.Router) {

				r.Get("/", h.partnerListCustomers)

				r.Post("/", h.partnerCreateCustomer)

				r.Route("/{id}", func(r chi.Router) {

					r.Get("/", h.partnerGetCustomer)

					r.Patch("/", h.partnerUpdateCustomer)

					r.Get("/phone-lines", h.partnerListCustomerPhoneLines)

				})

			})



			r.Route("/partner/phone-lines", func(r chi.Router) {

				r.Get("/", h.partnerListPhoneLines)

			})



			r.Route("/partner/phone-line-operation-requests", func(r chi.Router) {

				r.Get("/", h.partnerListLineOperationRequests)

				r.Post("/", h.partnerCreateLineOperationRequest)

			})



			r.Get("/partner/financial/summary", h.partnerGetFinancialSummary)

			r.Get("/partner/sales", h.partnerListSales)



			r.Get("/partner/contract-templates", h.partnerListContractTemplates)



			r.Route("/partner/commercial-sales", func(r chi.Router) {

				r.Get("/", h.partnerListCommercialSales)

				r.Post("/", h.partnerCreateCommercialSale)

				r.Route("/{id}", func(r chi.Router) {

					r.Get("/", h.partnerGetCommercialSale)

					r.Patch("/", h.partnerUpdateCommercialSale)

					r.Post("/confirm", h.partnerConfirmCommercialSale)

					r.Post("/cancel", h.partnerCancelCommercialSale)

					r.Post("/items", h.partnerAddSaleLineItem)

					r.Delete("/items/{itemId}", h.partnerDeleteSaleLineItem)

				})

			})

		})

	})

}



func (h *Handler) getDashboardStats(w http.ResponseWriter, r *http.Request) {

	stats, err := h.Svc.GetDashboardStats(r.Context())

	if err != nil {

		httputil.HandleServiceError(w, err)

		return

	}

	httputil.WriteJSON(w, http.StatusOK, stats)

}



func (h *Handler) postPresignedUpload(w http.ResponseWriter, r *http.Request) {

	if h.Presigned == nil {

		httputil.WriteFail(w, http.StatusServiceUnavailable, notifications.ObjectStorageUnavailable)

		return

	}

	var input models.CreatePresignedUploadURLInput

	if err := decodeJSON(r, &input); err != nil {

		httputil.WriteFail(w, http.StatusBadRequest, notifications.N("REQUEST_VALIDATION", "Invalid request body"))

		return

	}

	result, err := h.Presigned.CreateUploadURL(r.Context(), input)

	if err != nil {

		httputil.HandleServiceError(w, err)

		return

	}

	httputil.WriteJSON(w, http.StatusOK, result)

}



func (h *Handler) postPresignedDownload(w http.ResponseWriter, r *http.Request) {

	if h.Presigned == nil {

		httputil.WriteFail(w, http.StatusServiceUnavailable, notifications.ObjectStorageUnavailable)

		return

	}

	var input models.CreatePresignedDownloadURLInput

	if err := decodeJSON(r, &input); err != nil {

		httputil.WriteFail(w, http.StatusBadRequest, notifications.N("REQUEST_VALIDATION", "Invalid request body"))

		return

	}

	result, err := h.Presigned.CreateDownloadURL(r.Context(), input)

	if err != nil {

		httputil.HandleServiceError(w, err)

		return

	}

	httputil.WriteJSON(w, http.StatusOK, result)

}



func queryParam(r *http.Request, key string) *string {

	v := r.URL.Query().Get(key)

	if v == "" {

		return nil

	}

	return &v

}


