data "google_project" "current" {
  project_id = var.project_id
}

resource "google_billing_budget" "chat_demo" {
  billing_account = var.billing_account
  display_name    = "chat-demo-budget"

  budget_filter {
    # Budget API requires the project NUMBER here, not the project ID.
    projects               = ["projects/${data.google_project.current.number}"]
    credit_types_treatment = "EXCLUDE_ALL_CREDITS"
  }

  amount {
    specified_amount {
      # currency_code intentionally omitted — it must match the billing account's
      # currency, so we let it default to that (e.g. CAD) instead of hardcoding USD.
      units = "1"
    }
  }

  threshold_rules {
    threshold_percent = 0.5
  }

  threshold_rules {
    threshold_percent = 1.0
  }
}
