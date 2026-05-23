resource "google_billing_budget" "chat_demo" {
  billing_account = var.billing_account
  display_name    = "chat-demo-budget"

  budget_filter {
    projects               = ["projects/${var.project_id}"]
    credit_types_treatment = "EXCLUDE_ALL_CREDITS"
  }

  amount {
    specified_amount {
      currency_code = "USD"
      units         = "1"
    }
  }

  threshold_rules {
    threshold_percent = 0.5
  }

  threshold_rules {
    threshold_percent = 1.0
  }
}
